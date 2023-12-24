package data

import (
	"context"
)

type QueryResponse struct {
	Data             []byte
	AsyncRequestMade bool
}

type DataSource interface {
	Query(ctx context.Context, queryName string, query string, variables map[string]interface{}) (*QueryResponse, error)
}
