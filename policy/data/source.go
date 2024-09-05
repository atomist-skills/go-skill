package data

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/data/proxy"
	"github.com/atomist-skills/go-skill/policy/data/query"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
)

type DataSource interface {
	GetQueryClient() query.QueryClient
	GetProxyClient() (*proxy.ProxyClient, error)

	GetImageVulnerabilities(ctx context.Context, evalCtx goals.GoalEvaluationContext, imageSbom types.SBOM) (*query.QueryResponse, []types.Package, map[string][]types.Vulnerability, error)
}
