package policy_handler

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/goals"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
)

func WithSyncQuery() Opt {
	return func(h *EventHandler) {
		h.dataSourceProviders = append(h.dataSourceProviders, getSyncDataSources)
	}
}

func getSyncDataSources(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
	gqlDs := data.NewSyncGraphqlDataSourceFromSkillRequest(ctx, req, evalMeta)

	return []data.DataSource{
		gqlDs,
	}, nil
}
