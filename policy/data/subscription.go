package data

import (
	"context"
	"fmt"
	"olympos.io/encoding/edn"
)

// SubscriptionDataSource allows querying of objects directly from the subscription results array.
// It is beneficial to use this data source when it is useful to query by name,
// possibly allowing earlier data sources to pick up the query first.
type SubscriptionDataSource struct {
	queryIndexes        map[string]int
	subscriptionResults [][]edn.RawMessage
}

func NewSubscriptionDataSource(queryIndexes map[string]int, subscriptionResults [][]edn.RawMessage) SubscriptionDataSource {
	return SubscriptionDataSource{
		queryIndexes:        queryIndexes,
		subscriptionResults: subscriptionResults,
	}
}

func (ds SubscriptionDataSource) Query(_ context.Context, queryName string, _ string, _ map[string]interface{}, output interface{}) (*QueryResponse, error) {
	ix, ok := ds.queryIndexes[queryName]
	if !ok {
		return nil, nil
	}

	if ix >= len(ds.subscriptionResults) {
		return nil, fmt.Errorf("can't get subscription query %s (index %d) from result length %d", queryName, ix, len(ds.subscriptionResults))
	}

	err := edn.Unmarshal(ds.subscriptionResults[0][ix], output)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{}, nil
}
