package policy_handler

import (
	"context"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
)

func withSubscription() Opt {
	return func(h *EventHandler) {
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getSubscriptionData)
	}
}

func getSubscriptionData(ctx context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, error) {
	if req.Event.Context.Subscription.Name == "" {
		return nil, skill.Configuration{}, nil
	}

	evalMeta := &goals.EvaluationMetadata{
		SubscriptionResult: req.Event.Context.Subscription.Result,
		SubscriptionTx:     req.Event.Context.Subscription.Metadata.Tx,
	}
	return evalMeta, req.Event.Context.Subscription.Configuration, nil
}
