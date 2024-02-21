package policy_handler

import (
	"context"
	"fmt"

	"github.com/atomist-skills/go-skill/policy/goals"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
)

// WithSyncQuery enables synchronous queries for the specific query names provided.
// Queries not named in this constructor will be ignored by the synchronous data source.
//
// Usage of synchronous queries should be avoided whenever possible.
func WithSyncQuery(queryNames []string) Opt {
	return func(h *EventHandler) {
		h.dataSourceProviders = append(h.dataSourceProviders, getSyncDataSources(queryNames))
	}
}

func getSyncDataSources(queryNames []string) dataSourceProvider {
	return func(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
		gqlDs, err := data.NewSyncGraphqlDataSource(ctx, req, queryNames)
		if err != nil {
			return nil, fmt.Errorf("unable to create data source: %w", err)
		}

		return []data.DataSource{
			gqlDs,
		}, nil
	}
}
