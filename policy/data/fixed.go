package data

import (
	"context"
)

// FixedDataSource returns static data from responses passed in at construction time
type FixedDataSource struct {
	data map[string][]byte
}

func NewFixedDataSource(data map[string][]byte) FixedDataSource {
	return FixedDataSource{
		data: data,
	}
}

func (ds FixedDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}) (*QueryResponse, error) {
	res, ok := ds.data[queryName]
	if !ok {
		return nil, nil
	}

	return &QueryResponse{Data: res}, nil
}
