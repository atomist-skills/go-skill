package policy_handler

import (
	"context"
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/storage"
	"github.com/atomist-skills/go-skill/util"
	"olympos.io/encoding/edn"
)

type (
	EvaluatorSelector func(ctx context.Context, req skill.RequestContext, goal goals.Goal, dataSource data.DataSource) (goals.GoalEvaluator, error)

	Handler interface {
		Start()
	}

	subscriptionProvider func(ctx context.Context, req skill.RequestContext) ([][]edn.RawMessage, skill.Configuration, error)
	dataSourceProvider   func(ctx context.Context, req skill.RequestContext) ([]data.DataSource, error)
	transactionFilter    func(ctx context.Context, req skill.RequestContext) bool

	EventHandler struct {
		// parameters
		evalSelector      EvaluatorSelector
		subscriptionNames []string

		// hooks used by opts
		subscriptionDataProviders []subscriptionProvider
		dataSourceProviders       []dataSourceProvider
		transactFilters           []transactionFilter
	}

	Opt func(handler *EventHandler)
)

var defaultOpts = []Opt{
	withAsync(),
	withSubscription(),
}

func NewPolicyEventHandler(subscriptionNames []string, evalSelector EvaluatorSelector, opts ...Opt) EventHandler {
	p := EventHandler{
		subscriptionNames: subscriptionNames,
		evalSelector:      evalSelector,
	}

	for _, o := range opts {
		o(&p)
	}
	for _, o := range defaultOpts {
		o(&p)
	}

	return p
}

func (h EventHandler) Start() {
	handlers := skill.Handlers{}
	for _, n := range h.subscriptionNames {
		handlers[n] = h.handle
	}

	skill.Start(handlers)
}

func (h EventHandler) handle(ctx context.Context, req skill.RequestContext) skill.Status {
	var (
		subscriptionResult [][]edn.RawMessage
		configuration      skill.Configuration
		err                error
	)
	for _, provider := range h.subscriptionDataProviders {
		subscriptionResult, configuration, err = provider(ctx, req)
		if err != nil {
			return skill.NewFailedStatus(fmt.Sprintf("failed to retrieve subscription result [%s]", err.Error()))
		}
		if subscriptionResult != nil {
			break
		}
	}

	if subscriptionResult == nil {
		return skill.NewFailedStatus(fmt.Sprintf("subscription result was not found"))
	}

	sources := []data.DataSource{}
	for _, provider := range h.dataSourceProviders {
		ds, err := provider(ctx, req)
		if err != nil {
			if err.Error() == "An unexpected error has occurred" {
				return skill.NewRetryableStatus(fmt.Sprintf("Failed to create data source [%s]", err.Error()))
			}
			return skill.NewFailedStatus(fmt.Sprintf("failed to create data source [%s]", err.Error()))
		}
		sources = append(sources, ds...)
	}

	dataSource := data.NewChainDataSource(sources...)

	return h.evaluate(ctx, req, dataSource, subscriptionResult, configuration)
}

func (h EventHandler) evaluate(ctx context.Context, req skill.RequestContext, dataSource data.DataSource, subscriptionResult [][]edn.RawMessage, configuration skill.Configuration) skill.Status {
	goalName := req.Event.Skill.Name

	cfg := configuration.Name
	params := configuration.Parameters

	paramValues := map[string]interface{}{}
	for _, p := range params {
		paramValues[p.Name] = p.Value
	}

	// atm-skill local appends the current user's name to the skill name
	// we can strip that suffix off before calling evalSelector to let it match on the original name
	goalDefName := goalName
	u, err := user.Current()
	if err == nil {
		goalDefName = strings.TrimSuffix(goalDefName, fmt.Sprintf("-%s", u.Username))
	}

	goal := goals.Goal{
		Definition:    goalDefName,
		Configuration: cfg,
		Args:          paramValues,
	}

	evaluator, err := h.evalSelector(ctx, req, goal, dataSource)
	if err != nil {
		req.Log.Errorf(err.Error())
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create goal evaluator: %s", err.Error()))
	}

	commonResults := util.Decode[goals.CommonSubscriptionQueryResult](subscriptionResult[0][0])
	digest := commonResults.ImageDigest

	req.Log.Infof("Evaluating goal %s for digest %s ", goalName, digest)
	evaluationTs := time.Now().UTC()

	goalResults, err := evaluator.EvaluateGoal(ctx, req, commonResults, subscriptionResult)
	if err != nil {
		req.Log.Errorf("Failed to evaluate goal %s for digest %s: %s", goal.Definition, digest, err.Error())
		return skill.NewFailedStatus("Failed to evaluate goal")
	}

	for _, f := range h.transactFilters {
		if !f(ctx, req) {
			// if not transacting, we return results as part of the skill result
			return skill.Status{
				State:       skill.Completed,
				Reason:      fmt.Sprintf("Goal %s evaluated", goalName),
				SyncRequest: goalResults,
			}
		}
	}

	return transact(
		ctx,
		req,
		configuration,
		goalName,
		digest,
		goal,
		subscriptionResult,
		evaluationTs,
		goalResults,
	)
}

func transact(
	ctx context.Context,
	req skill.RequestContext,
	configuration skill.Configuration,
	goalName string,
	digest string,
	goal goals.Goal,
	subscriptionResult [][]edn.RawMessage,
	evaluationTs time.Time,
	goalResults []goals.GoalEvaluationQueryResult,
) skill.Status {
	storageTuple := util.Decode[[]string](subscriptionResult[0][1])
	storageId := storageTuple[0]
	configHash := storageTuple[1]

	if goalResults == nil {
		req.Log.Infof("goal %s returned no data for digest %s, skipping storing results", goal.Definition, digest)
		return skill.NewCompletedStatus(fmt.Sprintf("Goal %s evaluated - no data found", goalName))
	}

	es, err := storage.NewEvaluationStorage(ctx)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create evaluation storage: %s", err.Error()))
	}

	configDiffer, configHash, err := goals.GoalConfigsDiffer(req.Log, configuration, digest, goal, configHash)
	if err != nil {
		req.Log.Errorf("Failed to check if config hash changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		configDiffer = true
	}

	differ, storageId, err := goals.GoalResultsDiffer(req.Log, goalResults, digest, goal, storageId)
	if err != nil {
		req.Log.Errorf("Failed to check if goal results changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		differ = true
	}

	if differ {
		if err := es.Store(ctx, goalResults, storageId, req.Event.Environment, req.Log); err != nil {
			return skill.NewFailedStatus(fmt.Sprintf("Failed to store evaluation results for digest %s: %s", digest, err.Error()))
		}
	}

	var entities []interface{}
	if differ || configDiffer {
		entity := goals.CreateEntitiesFromResults(goalResults, goal.Definition, goal.Configuration, digest, storageId, configHash, evaluationTs)
		entities = append(entities, entity)
	}

	if len(entities) > 0 {
		err = req.NewTransaction().AddEntities(entities...).Transact()
		if err != nil {
			req.Log.Errorf(err.Error())
		}
		req.Log.Info("Goal results transacted")
	} else {
		req.Log.Info("No goal results to transact")
	}

	return skill.NewCompletedStatus(fmt.Sprintf("Goal %s evaluated", goalName))
}
