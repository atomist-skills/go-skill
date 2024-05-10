package query

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/goals"
)

type QueryResponse struct {
	AsyncRequestMade bool
}

type QueryClient interface {
	Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error)
}

func GqlContext(ctx goals.GoalEvaluationContext) map[string]interface{} {
	result := map[string]interface{}{}

	if ctx.TeamId != "" {
		result["teamId"] = ctx.TeamId
	}

	if ctx.Organization != "" {
		result["organization"] = ctx.Organization
	}

	return result
}
