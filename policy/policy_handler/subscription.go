package policy_handler

import (
	"context"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"olympos.io/encoding/edn"
)

func WithSubscription() Opt {
	return func(h *EventHandler) {
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, getSyncDataSources)
	}
}

func getSubscriptionData(ctx context.Context, req skill.RequestContext) ([][]edn.RawMessage, skill.Configuration, error) {
	if req.Event.Context.Subscription.Name == "" {
		return nil, skill.Configuration{}, nil
	}

	return req.Event.Context.Subscription.Result, req.Event.Context.Subscription.Configuration, nil
}

func getSyncDataSources(ctx context.Context, req skill.RequestContext) ([]data.DataSource, error) {
	gqlDs, err := data.NewSyncGraphqlDataSource(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create data source: %w", err)
	}

	return []data.DataSource{
		gqlDs,
	}, nil
}
