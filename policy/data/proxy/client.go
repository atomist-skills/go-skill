package proxy

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
)

type ProxyClient interface {
	Evaluate(ctx context.Context, organization, teamId, url string, sbom *types.SBOM, args map[string]interface{}) (goals.EvaluationResult, error)
}
