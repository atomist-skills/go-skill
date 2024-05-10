package query

import (
	"context"
)

type FixedQueryClientUnmarshaler func(data []byte, output interface{}) error

// FixedQueryClient returns static data from responses passed in at construction time
type FixedQueryClient struct {
	unmarshaler FixedQueryClientUnmarshaler
	data        map[string][]byte
}

func NewFixedQueryClient(unmarshaler FixedQueryClientUnmarshaler, data map[string][]byte) FixedQueryClient {
	return FixedQueryClient{
		unmarshaler: unmarshaler,
		data:        data,
	}
}

func (ds FixedQueryClient) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	res, ok := ds.data[queryName]
	if !ok {
		return nil, nil
	}

	err := ds.unmarshaler(res, output)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{}, nil
}
