package policy_handler

import (
	"context"
	"encoding/json"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
)

const eventNameLocalEval = "evaluate_goals_locally"

func WithLocal() Opt {
	return func(h *EventHandler) {
		h.subscriptionNames = append(h.subscriptionNames, eventNameLocalEval)
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getLocalSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, buildLocalDataSources)
		h.transactFilters = append(h.transactFilters, shouldTransactLocal)
	}
}

func getLocalSubscriptionData(ctx context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return nil, skill.Configuration{}, nil
	}

	return &goals.EvaluationMetadata{}, req.Event.Context.Subscription.Configuration, nil
}

func buildLocalDataSources(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return []data.DataSource{}, nil
	}

	metaFixedData := map[string][]byte{}
	for k, v := range req.Event.Context.SyncRequest.Metadata {
		metaFixedData[string(k)] = v
	}
	fixedDs := data.NewFixedDataSource(json.Unmarshal, metaFixedData)

	return []data.DataSource{
		fixedDs,
	}, nil
}

func shouldTransactLocal(ctx context.Context, req skill.RequestContext) bool {
	return req.Event.Context.SyncRequest.Name != eventNameLocalEval
}
