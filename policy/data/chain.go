package data

import (
	"context"
	"fmt"
)

// ChainDataSource is a wrapper DataSource that takes a list of other DataSources
// and returns query results from the first applicable downstream source
type ChainDataSource struct {
	links []DataSource
}

func NewChainDataSource(links ...DataSource) *ChainDataSource {
	return &ChainDataSource{
		links: links,
	}
}

func (ds ChainDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}) (*QueryResponse, error) {
	for _, l := range ds.links {
		res, err := l.Query(ctx, queryName, query, variables)
		if res != nil || err != nil {
			return res, err
		}
	}

	return nil, fmt.Errorf("no DataSource was available to process query %s", queryName)
}
