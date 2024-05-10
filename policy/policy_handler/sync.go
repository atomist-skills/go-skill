package policy_handler

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/data/query"
	"github.com/atomist-skills/go-skill/policy/goals"

	"github.com/atomist-skills/go-skill"
)

func WithSyncQuery() Opt {
	return func(h *EventHandler) {
		h.queryClientProviders = append(h.queryClientProviders, getSyncQueryClients)
	}
}

func getSyncQueryClients(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]query.QueryClient, error) {
	gqlDs := query.NewSyncGraphqlQueryClientFromSkillRequest(ctx, req, evalMeta)

	return []query.QueryClient{
		gqlDs,
	}, nil
}
