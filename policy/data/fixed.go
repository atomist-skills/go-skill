package data

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/data/proxy"
	"github.com/atomist-skills/go-skill/policy/data/query"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
)

type vulnerabilityFetcher func(ctx context.Context, evalCtx goals.GoalEvaluationContext, imageSbom types.SBOM) (*query.QueryResponse, []types.Package, map[string][]types.Vulnerability, error)

type fixedDataSource struct {
	jynxGQLClient   query.QueryClient
	proxyClient     proxy.ProxyClient
	vulnerabilities vulnerabilityFetcher
}

func NewFixedDataSource(jynxGQLClient query.QueryClient, proxyClient proxy.ProxyClient, vulnerabilities vulnerabilityFetcher) DataSource {
	return &fixedDataSource{
		jynxGQLClient:   jynxGQLClient,
		proxyClient:     proxyClient,
		vulnerabilities: vulnerabilities,
	}
}

func (ds *fixedDataSource) GetQueryClient() query.QueryClient {
	return ds.jynxGQLClient
}

func (ds *fixedDataSource) GetProxyClient() (proxy.ProxyClient, error) {
	return ds.proxyClient, nil
}

func (ds *fixedDataSource) GetImageVulnerabilities(ctx context.Context, evalCtx goals.GoalEvaluationContext, imageSbom types.SBOM) (*query.QueryResponse, []types.Package, map[string][]types.Vulnerability, error) {
	if ds.vulnerabilities != nil {
		return ds.vulnerabilities(ctx, evalCtx, imageSbom)
	}

	return &query.QueryResponse{}, []types.Package{}, map[string][]types.Vulnerability{}, nil
}
