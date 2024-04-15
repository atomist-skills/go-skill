package cache

import (
	"context"
	"encoding/json"
)

type QueryCache interface {
	Read(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error)
	Write(ctx context.Context, query string, variables map[string]interface{}, res []byte) error
}

type cacheKey struct {
	query     string
	variables string
}

type SimpleQueryCache struct {
	cache map[cacheKey][]byte
}

func NewQueryCache() SimpleQueryCache {
	return SimpleQueryCache{
		cache: make(map[cacheKey][]byte),
	}
}

func (d SimpleQueryCache) Read(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}

	key := cacheKey{query: query, variables: string(variablesJSON)}
	if result, ok := d.cache[key]; ok {
		return result, nil
	}

	return nil, nil
}

func (d SimpleQueryCache) Write(ctx context.Context, query string, variables map[string]interface{}, res []byte) error {
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return err
	}

	key := cacheKey{query: query, variables: string(variablesJSON)}
	d.cache[key] = res

	return nil
}
