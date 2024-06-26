package data

import (
	"fmt"

	"github.com/atomist-skills/go-skill/policy/data/proxy"
	"github.com/atomist-skills/go-skill/policy/data/query"
)

type DataSource struct {
	jynxGQLClient query.QueryClient
	proxyClient   *proxy.ProxyClient
}

func NewDataSource(graphQLClient query.QueryClient, proxyClient *proxy.ProxyClient) DataSource {
	return DataSource{
		jynxGQLClient: graphQLClient,
		proxyClient:   proxyClient,
	}
}

func (ds *DataSource) GetQueryClient() query.QueryClient {
	return ds.jynxGQLClient
}

func (ds *DataSource) GetProxyClient() (*proxy.ProxyClient, error) {
	if ds.proxyClient == nil {
		return nil, fmt.Errorf("no proxy client is configured")
	}

	return ds.proxyClient, nil
}
