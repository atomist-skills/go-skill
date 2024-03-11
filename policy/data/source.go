package data

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/goals"
)

type QueryResponse struct {
	AsyncRequestMade bool
}

type DataSource interface {
	Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error)
}

func GqlContext(ctx goals.GoalEvaluationContext) map[string]interface{} {
	return map[string]interface{}{
		"teamId":       ctx.TeamId,
		"organization": ctx.Organization,
	}
}
