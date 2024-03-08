package data

import (
	"context"

	"olympos.io/encoding/edn"
)

// SubscriptionDataSource allows querying of objects directly from the subscription results array.
// It is beneficial to use this data source when it is useful to query by name,
// possibly allowing earlier data sources to pick up the query first.
type SubscriptionDataSource struct {
	queryIndexes        map[string]int
	subscriptionResults []map[edn.Keyword]edn.RawMessage
}

func NewSubscriptionDataSource(queryIndexes map[string]int, subscriptionResults []map[edn.Keyword]edn.RawMessage) SubscriptionDataSource {
	return SubscriptionDataSource{
		queryIndexes:        queryIndexes,
		subscriptionResults: subscriptionResults,
	}
}

func (ds SubscriptionDataSource) Query(_ context.Context, queryName string, _ string, _ map[string]interface{}, output interface{}) (*QueryResponse, error) {
	if len(ds.subscriptionResults) == 0 {
		return nil, nil
	}

	result, ok := ds.subscriptionResults[0][edn.Keyword(queryName)]
	if !ok {
		return nil, nil
	}

	err := edn.Unmarshal(result, output)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{}, nil
}
