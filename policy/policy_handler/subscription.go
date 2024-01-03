package policy_handler

import (
	"context"
	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

func withSubscription() Opt {
	return func(h *EventHandler) {
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getSubscriptionData)
	}
}

func getSubscriptionData(ctx context.Context, req skill.RequestContext) ([][]edn.RawMessage, skill.Configuration, error) {
	if req.Event.Context.Subscription.Name == "" {
		return nil, skill.Configuration{}, nil
	}

	return req.Event.Context.Subscription.Result, req.Event.Context.Subscription.Configuration, nil
}
