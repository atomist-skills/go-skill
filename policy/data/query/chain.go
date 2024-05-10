package query

import (
	"context"
	"fmt"
)

// ChainQueryClient is a wrapper QueryClient that takes a list of other QueryClients
// and returns query results from the first applicable downstream source
type ChainQueryClient struct {
	links []QueryClient
}

func NewChainQueryClient(links ...QueryClient) *ChainQueryClient {
	return &ChainQueryClient{
		links: links,
	}
}

func (ds ChainQueryClient) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	for _, l := range ds.links {
		res, err := l.Query(ctx, queryName, query, variables, output)
		if res != nil || err != nil {
			return res, err
		}
	}

	return nil, fmt.Errorf("no QueryClient was available to process query %s", queryName)
}
