package cache

import (
	"context"
	"encoding/json"
	"sync"
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
	mu    sync.Mutex
	cache map[cacheKey][]byte
}

func NewQueryCache() SimpleQueryCache {
	return SimpleQueryCache{
		cache: make(map[cacheKey][]byte),
	}
}

func getKey(query string, variables map[string]interface{}) (cacheKey, error) {
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return cacheKey{}, err
	}

	return cacheKey{query: query, variables: string(variablesJSON)}, nil
}

func (d *SimpleQueryCache) Read(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	key, err := getKey(query, variables)
	if err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if result, ok := d.cache[key]; ok {
		return result, nil
	}

	return nil, nil
}

func (d *SimpleQueryCache) Write(ctx context.Context, query string, variables map[string]interface{}, res []byte) error {
	key, err := getKey(query, variables)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache[key] = res

	return nil
}
