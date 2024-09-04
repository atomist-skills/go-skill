package query

import (
	"context"
)

type QueryResponse struct {
	AsyncRequestMade bool
}

type QueryClient interface {
	Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error)
}
