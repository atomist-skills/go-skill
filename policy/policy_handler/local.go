package policy_handler

import (
	"context"
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"olympos.io/encoding/edn"
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

func getLocalSubscriptionData(ctx context.Context, req skill.RequestContext) ([][]edn.RawMessage, skill.Configuration, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return nil, skill.Configuration{}, nil
	}

	return [][]edn.RawMessage{}, req.Event.Context.Subscription.Configuration, nil
}

func buildLocalDataSources(ctx context.Context, req skill.RequestContext) ([]data.DataSource, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return []data.DataSource{}, nil
	}

	metaFixedData := map[string][]byte{}
	for k, v := range req.Event.Context.SyncRequest.Metadata {
		metaFixedData[string(k)] = v
	}
	fixedDs := data.NewFixedDataSource(metaFixedData)

	return []data.DataSource{
		fixedDs,
	}, nil
}

func shouldTransactLocal(ctx context.Context, req skill.RequestContext) bool {
	return req.Event.Context.SyncRequest.Name != eventNameLocalEval
}
