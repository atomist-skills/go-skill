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
	"github.com/atomist-skills/go-skill/policy/transact"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/atomist-skills/go-skill/util"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

type (
	EvaluatorSelector func(ctx context.Context, goal goals.Goal, dataSource data.DataSource) (goals.GoalEvaluator, error)

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

	evaluator, err := h.evalSelector(ctx, goal, dataSource)
	if err != nil {
		req.Log.Errorf(err.Error())
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create goal evaluator: %s", err.Error()))
	}

	digest := sbom.Source.Image.Digest

	req.Log.Infof("Evaluating goal %s for digest %s ", goalName, digest)
	evaluationTs := time.Now().UTC()

	evalContext := goals.GoalEvaluationContext{
		Log:          req.Log,
		TeamId:       req.Event.WorkspaceId,
		Organization: req.Event.Organization,
		Goal:         goal,
	}

	evaluationResult, err := evaluator.EvaluateGoal(ctx, evalContext, sbom, subscriptionResult)
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

	storageTuple := util.Decode[[]string](subscriptionResult[0]["previous"])

	if len(storageTuple) != 2 {
		req.Log.Error("could not find previous result in subscription result")
		return skill.Status{
			State:  skill.Failed,
			Reason: "could not find previous result in subscription result",
		}
	}

	previousResult := transact.PreviousResult{
		StorageId:  storageTuple[0],
		ConfigHash: storageTuple[1],
	}

	err = transact.TransactPolicyResult(
		ctx,
		evalContext,
		configuration,
		digest,
		&previousResult,
		evaluationTs,
		goalResults,
		tx,
		req.NewTransaction,
	)

	if err != nil {
		req.Log.Errorf("Failed to transact goal results: %s", err.Error())
		return skill.NewFailedStatus(fmt.Sprintf("Failed to transact goal results: %s", err.Error()))
	}

	return skill.NewCompletedStatus(fmt.Sprintf("Goal %s evaluated", goalName))
}

type intotoStatement struct {
	intoto.StatementHeader
	Predicate json.RawMessage `json:"predicate"`
}
