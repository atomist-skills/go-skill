package policy_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/user"
	"strings"
	"time"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/storage"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/atomist-skills/go-skill/util"
	"olympos.io/encoding/edn"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

type (
	EvaluatorSelector func(ctx context.Context, req skill.RequestContext, goal goals.Goal, dataSource data.DataSource) (goals.GoalEvaluator, error)

	evalInputProvider  func(ctx context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, *types.SBOM, error)
	dataSourceProvider func(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error)
	transactionFilter  func(ctx context.Context, req skill.RequestContext) bool

	EventHandler struct {
		// parameters
		evalSelector      EvaluatorSelector
		subscriptionNames []string

		// hooks used by opts
		evalInputProviders  []evalInputProvider
		dataSourceProviders []dataSourceProvider
		transactFilters     []transactionFilter
	}

	Opt func(handler *EventHandler)
)

var defaultOpts = []Opt{
	WithAsync(),
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

func (h EventHandler) createSkillHandlers() skill.Handlers {
	handlers := skill.Handlers{}
	for _, n := range h.subscriptionNames {
		handlers[n] = h.handle
	}

	return handlers
}

func (h EventHandler) Start() {
	handlers := h.createSkillHandlers()

	skill.Start(handlers)
}

func (h EventHandler) CreateHttpHandler() func(http.ResponseWriter, *http.Request) {
	handlers := h.createSkillHandlers()

	return skill.CreateHttpHandler(handlers)
}

func (h EventHandler) ExecuteSyncRequest(ctx context.Context, req skill.RequestContext) ([]goals.GoalEvaluationQueryResult, error) {
	handlers := h.createSkillHandlers()

	syncHandler, ok := handlers[eventNameLocalEval]
	if !ok {
		return nil, fmt.Errorf("no handler for sync request")
	}

	result := syncHandler(ctx, req)

	if result.State != skill.Completed {
		return nil, fmt.Errorf("sync request did not complete successfully [%s]", result.Reason)
	}

	return result.SyncRequest.([]goals.GoalEvaluationQueryResult), nil
}

func (h EventHandler) handle(ctx context.Context, req skill.RequestContext) skill.Status {
	var (
		evaluationMetadata *goals.EvaluationMetadata
		configuration      skill.Configuration
		sbom               *types.SBOM
		err                error
	)
	for _, provider := range h.evalInputProviders {
		evaluationMetadata, configuration, sbom, err = provider(ctx, req)
		if err != nil {
			return skill.NewFailedStatus(fmt.Sprintf("failed to retrieve subscription result [%s]", err.Error()))
		}
		if evaluationMetadata != nil {
			break
		}
	}

	if evaluationMetadata == nil {
		return skill.NewFailedStatus("subscription result was not found")
	}

	sources := []data.DataSource{}
	for _, provider := range h.dataSourceProviders {
		ds, err := provider(ctx, req, *evaluationMetadata)
		if err != nil {
			if retryableError, ok := err.(types.RetryableExecutionError); ok {
				return skill.NewRetryableStatus(fmt.Sprintf("Failed to create data source [%s]", retryableError.Error()))
			}
			return skill.NewFailedStatus(fmt.Sprintf("failed to create data source [%s]", err.Error()))
		}
		sources = append(sources, ds...)
	}

	dataSource := data.NewChainDataSource(sources...)

	return h.evaluate(ctx, req, dataSource, *evaluationMetadata, *sbom, configuration)
}

func (h EventHandler) evaluate(ctx context.Context, req skill.RequestContext, dataSource data.DataSource, evaluationMetadata goals.EvaluationMetadata, sbom types.SBOM, configuration skill.Configuration) skill.Status {
	goalName := req.Event.Skill.Name
	tx := evaluationMetadata.SubscriptionTx
	subscriptionResult := evaluationMetadata.SubscriptionResult

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

	if err != nil {
		req.Log.Errorf(err.Error())
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create sbom from subscription: %s", err.Error()))
	}
	digest := sbom.Source.Image.Digest

	req.Log.Infof("Evaluating goal %s for digest %s ", goalName, digest)
	evaluationTs := time.Now().UTC()

	evaluationResult, err := evaluator.EvaluateGoal(ctx, req, sbom)
	if err != nil {
		req.Log.Errorf("Failed to evaluate goal %s for digest %s: %s", goal.Definition, digest, err.Error())
		return skill.NewFailedStatus("Failed to evaluate goal")
	}

	if !evaluationResult.EvaluationCompleted {
		req.Log.Info("evaluation incomplete")
		return skill.NewCompletedStatus("Evaluation incomplete")
	}

	goalResults := evaluationResult.Result

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
		tx,
	)
}

type intotoStatement struct {
	intoto.StatementHeader
	Predicate json.RawMessage `json:"predicate"`
}

func transact(
	ctx context.Context,
	req skill.RequestContext,
	configuration skill.Configuration,
	goalName string,
	digest string,
	goal goals.Goal,
	subscriptionResult []map[edn.Keyword]edn.RawMessage,
	evaluationTs time.Time,
	goalResults []goals.GoalEvaluationQueryResult,
	tx int64,
) skill.Status {
	storageTuple := util.Decode[[]string](subscriptionResult[0]["previous"])
	previousStorageId := storageTuple[0]
	previousConfigHash := storageTuple[1]

	if goalResults == nil {
		req.Log.Infof("goal %s returned no data for digest %s", goal.Definition, digest)
	}

	es, err := storage.NewEvaluationStorage(ctx)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create evaluation storage: %s", err.Error()))
	}

	configDiffer, configHash, err := goals.GoalConfigsDiffer(req.Log, configuration, digest, goal, previousConfigHash)
	if err != nil {
		req.Log.Errorf("Failed to check if config hash changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		configDiffer = true
	}

	differ, storageId, err := goals.GoalResultsDiffer(req.Log, goalResults, digest, goal, previousStorageId)
	if err != nil {
		req.Log.Errorf("Failed to check if goal results changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		differ = true
	}

	if differ && goalResults != nil {
		if err := es.Store(ctx, goalResults, storageId, req.Event.Environment, req.Log); err != nil {
			return skill.NewFailedStatus(fmt.Sprintf("Failed to store evaluation results for digest %s: %s", digest, err.Error()))
		}
	}

	var entities []interface{}
	if differ || configDiffer {
		shouldRetract := previousStorageId != "no-data" && previousStorageId != "n/a" && storageId == "no-data"
		entity := goals.CreateEntitiesFromResults(goalResults, goal.Definition, goal.Configuration, digest, storageId, configHash, evaluationTs, tx, shouldRetract)
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
