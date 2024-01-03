package data

import (
	"context"

	"github.com/atomist-skills/go-skill"
)

type QueryResponse struct {
	AsyncRequestMade bool
}

type DataSource interface {
	Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error)
}

func GqlContext(req skill.RequestContext) map[string]interface{} {
	return map[string]interface{}{
		"teamId":       req.Event.WorkspaceId,
		"organization": req.Event.Organization,
	}
}
