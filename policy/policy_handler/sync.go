package policy_handler

import (
	"context"
	"fmt"
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
)

func WithSyncQuery() Opt {
	return func(h *EventHandler) {
		h.dataSourceProviders = append(h.dataSourceProviders, getSyncDataSources)
	}
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
