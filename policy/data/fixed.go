package data

import (
	"context"
)

type FixedDataSourceUnmarshaler func(data []byte, output interface{}) error

// FixedDataSource returns static data from responses passed in at construction time
type FixedDataSource struct {
	unmarshaler FixedDataSourceUnmarshaler
	data        map[string][]byte
}

func NewFixedDataSource(unmarshaler FixedDataSourceUnmarshaler, data map[string][]byte) FixedDataSource {
	return FixedDataSource{
		unmarshaler: unmarshaler,
		data:        data,
	}
}

func (ds FixedDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
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
