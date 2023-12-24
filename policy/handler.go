package policy

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"time"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/evaluators"
	"github.com/atomist-skills/go-skill/policy/evaluators/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/query"
	"github.com/atomist-skills/go-skill/policy/storage"
	"github.com/atomist-skills/go-skill/util"
	"olympos.io/encoding/edn"
)

type (
	EvaluatorSelector func(ctx context.Context, req skill.RequestContext, goal *goals.Goal, dataSource data.DataSource) (evaluators.GoalEvaluator, error)

	Handler interface {
		Start()
	}

	EventHandler struct {
		enableAsync       bool
		enableLocalEval   bool
		subscriptionNames []string
		evalSelector      EvaluatorSelector
	}

	Opt func(handler *EventHandler) error
)

func NewPolicyEventHandler(subscriptionNames []string, evalSelector EvaluatorSelector, opts ...Opt) (*EventHandler, error) {
	p := &EventHandler{
		subscriptionNames: subscriptionNames,
		evalSelector:      evalSelector,
	}

	for _, o := range opts {
		err := o(p)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func WithAsyncQuerySupport(p *EventHandler) error {
	p.enableAsync = true
	return nil
}

func WithLocalEvalSupport(p *EventHandler) error {
	p.enableLocalEval = true
	return nil
}

func (p *EventHandler) Start() {
	handlers := skill.Handlers{}
	for _, n := range p.subscriptionNames {
		handlers[n] = p.handle
	}

	if p.enableAsync {
		handlers["async-query-packages"] = p.handleAsync
		handlers["async-query-image-details"] = p.handleAsync
	}

	if p.enableLocalEval {
		handlers["evaluate_goals_locally"] = p.handleLocal
	}

	skill.Start(handlers)
}

// handleLocal runs the goal evaluation locally and returns the results without transacting them.
func (p *EventHandler) handleLocal(ctx context.Context, req skill.RequestContext) skill.Status {
	goalName := req.Event.Skill.Name

	cfg := req.Event.Context.SyncRequest.Configuration.Name
	params := req.Event.Context.SyncRequest.Configuration.Parameters

	values := map[string]interface{}{}
	for _, p := range params {
		values[p.Name] = p.Value
	}

	if _, ok := values["definitionName"]; !ok {
		return skill.NewFailedStatus("Missing definition name in policy skill configuration")
	}

	goal := goals.Goal{
		Definition:    values["definitionName"].(string),
		Configuration: cfg,
		Args:          values,
	}

	metaPkgs := util.Decode[[]data.MetadataPackage](req.Event.Context.SyncRequest.Metadata["packages"])

	digest := "localDigest"

	subscriptionResult := query.CommonSubscriptionQueryResult{
		ImageDigest: digest,
	}

	dataSource, err := data.NewQueryDataSource(ctx, req,
		data.WithFixedPackageList(map[string][]data.MetadataPackage{
			digest: metaPkgs,
		}),
	)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("unable to create data source: %s", err.Error()))
	}

	evaluator, err := p.evalSelector(ctx, req, &goal, dataSource)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("unable to create evaluator: %s", err.Error()))
	}

	if evaluator.GetFlags()&evaluators.EVAL_SKIP_LOCAL != 0 {
		return skill.NewCompletedStatus("Skipped eval due to EVAL_SKIP_LOCAL")
	}

	results, err := evaluator.EvaluateGoal(ctx, req, subscriptionResult, [][]edn.RawMessage{})
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Error evaluating goal %s", err.Error()))
	}

	return skill.Status{
		State:       skill.Completed,
		Reason:      fmt.Sprintf("Goal %s evaluated", goalName),
		SyncRequest: results,
	}
}

func (p *EventHandler) handleAsync(ctx context.Context, req skill.RequestContext) skill.Status {
	asyncClient := query.NewAsyncQueryClient(req.Log, req.Event.Token, req.Event.Context.AsyncQueryResult.Metadata)

	metadata := req.Event.Context.AsyncQueryResult.Metadata
	encoded, err := b64.StdEncoding.DecodeString(metadata)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to decode async metadata [%s]", err.Error()))
	}
	var subscriptionResult [][]edn.RawMessage

	err = edn.Unmarshal(encoded, &subscriptionResult)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to unmarshal async metadata [%s]", err.Error()))
	}

	dataSource, err := data.NewQueryDataSource(ctx, req,
		data.WithAsyncQueryResult(req.Event.Context.AsyncQueryResult.Name, req.Event.Context.AsyncQueryResult.Result),
		data.WithAsyncClient(asyncClient))
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create data source [%s]", err.Error()))
	}

	return p.evaluateGoalWithData(ctx, req, dataSource, subscriptionResult, req.Event.Context.AsyncQueryResult.Configuration)
}

// EvaluateGoals runs the goal evaluation and returns the results after transacting them.
func (p *EventHandler) handle(ctx context.Context, req skill.RequestContext) skill.Status {
	edn, err := edn.Marshal(req.Event.Context.Subscription.Result)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to marshal metadata [%s]", err.Error()))
	}

	encodedMetadata := b64.StdEncoding.EncodeToString(edn)

	asyncClient := query.NewAsyncQueryClient(req.Log, req.Event.Token, encodedMetadata)

	dataSource, err := data.NewQueryDataSource(ctx, req,
		data.WithAsyncClient(asyncClient))
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create data source [%s]", err.Error()))
	}

	return p.evaluateGoalWithData(ctx, req, dataSource, req.Event.Context.Subscription.Result, req.Event.Context.Subscription.Configuration)
}

func (p *EventHandler) evaluateGoalWithData(ctx context.Context, req skill.RequestContext, dataSource data.DataSource, subscriptionResult [][]edn.RawMessage, configuration skill.Configuration) skill.Status {
	goalName := req.Event.Skill.Name

	cfg := configuration.Name
	params := configuration.Parameters

	values := map[string]interface{}{}
	for _, p := range params {
		values[p.Name] = p.Value
	}

	if _, ok := values["definitionName"]; !ok {
		return skill.NewFailedStatus("Missing definition name in policy skill configuration")
	}

	goal := goals.Goal{
		Definition:    values["definitionName"].(string),
		Configuration: cfg,
		Args:          values,
	}

	evaluator, err := p.evalSelector(ctx, req, &goal, dataSource)
	if err != nil {
		req.Log.Errorf(err.Error())
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create goal evaluator: %s", err.Error()))
	}

	es, err := storage.NewEvaluationStorage(ctx)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create evaluation storage: %s", err.Error()))
	}

	var entities []interface{}
	var commonResults query.CommonSubscriptionQueryResult
	var previousStorageId string
	var previousConfigHash string

	// Find correct storageId and configHash from subscription result by only returning n/a if that is the only option
	for _, result := range subscriptionResult {
		storageTuple := util.Decode[[]string](result[1])
		if previousStorageId == "" || previousStorageId == "n/a" {
			previousStorageId = storageTuple[0]
		}
		if previousConfigHash == "" || previousConfigHash == "n/a" {
			previousConfigHash = storageTuple[1]
		}
	}

	subscriptionResults := [][]edn.RawMessage{}
	for _, result := range subscriptionResult {
		// Filter out duplicate results if we have real storage id and n/a
		storageTuple := util.Decode[[]string](result[1])
		resultPreviousStorageId := storageTuple[0]
		if resultPreviousStorageId == previousStorageId {
			subscriptionResults = append(subscriptionResults, result)
		}

		commonResults = util.Decode[query.CommonSubscriptionQueryResult](result[0])
	}

	digest := commonResults.ImageDigest
	req.Log.Infof("Evaluating goal %s for digest %s ", goalName, digest)
	evaluationTs := time.Now().UTC()

	qr, err := evaluator.EvaluateGoal(ctx, req, commonResults, subscriptionResults)
	if err != nil {
		req.Log.Errorf("Failed to evaluate goal %s for digest %s: %s", goal.Definition, digest, err.Error())
		return skill.NewFailedStatus("Failed to evaluate goal")
	}

	if qr == nil {
		req.Log.Infof("goal %s returned no data for digest %s, skipping storing results", goal.Definition, digest)
		return skill.NewCompletedStatus(fmt.Sprintf("Goal %s evaluated - no data found", goalName))
	}

	configDiffer, configHash, err := evaluators.GoalConfigsDiffer(req.Log, configuration, digest, goal, previousConfigHash)
	if err != nil {
		req.Log.Errorf("Failed to check if config hash changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		configDiffer = true
	}

	differ, storageId, err := evaluators.GoalResultsDiffer(req.Log, qr, digest, goal, previousStorageId)
	if err != nil {
		req.Log.Errorf("Failed to check if goal results changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		differ = true
	}

	if differ {
		if err := es.Store(ctx, qr, storageId, req.Event.Environment, req.Log); err != nil {
			return skill.NewFailedStatus(fmt.Sprintf("Failed to store evaluation results for digest %s: %s", digest, err.Error()))
		}
	}

	if differ || configDiffer {
		entity := goals.CreateEntitiesFromResults(qr, goal.Definition, goal.Configuration, digest, storageId, configHash, evaluationTs)
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
