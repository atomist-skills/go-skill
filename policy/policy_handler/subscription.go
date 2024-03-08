package policy_handler

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"

	"github.com/atomist-skills/go-skill"
)

func withSubscription() Opt {
	return func(h *EventHandler) {
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getSubscriptionData)
	}
}

func getSubscriptionData(_ context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, error) {
	if req.Event.Context.Subscription.Name == "" {
		return nil, skill.Configuration{}, nil
	}

	evalMeta := &goals.EvaluationMetadata{
		SubscriptionResult: req.Event.Context.Subscription.GetResultInMapForm(),
		SubscriptionTx:     req.Event.Context.Subscription.Metadata.Tx,
	}
	return evalMeta, req.Event.Context.Subscription.Configuration, nil
}

func WithSubscriptionDataSource(queryIndexes map[string]int) Opt {
	return func(h *EventHandler) {
		h.dataSourceProviders = append(h.dataSourceProviders, buildSubscriptionDataSource(queryIndexes))
	}
}

func buildSubscriptionDataSource(queryIndexes map[string]int) dataSourceProvider {
	return func(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
		return []data.DataSource{
			data.NewSubscriptionDataSource(queryIndexes, evalMeta.SubscriptionResult),
		}, nil
	}
}
