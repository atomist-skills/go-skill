package policy_handler

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"olympos.io/encoding/edn"
)

const eventNameAsyncQuery = data.AsyncQueryName // these must match for the event handler to be registered

// WithAsync will enable async graphql queries for the EventHandler.
// When used, data.QueryResponse#AsyncRequestMade will be true when performed asynchronously.
// If that flag is set, the policy evaluator should terminate early with no results.
// It will be automatically retried once the async query results are returned.
func WithAsync() Opt {
	return func(h *EventHandler) {
		h.subscriptionNames = append(h.subscriptionNames, eventNameAsyncQuery)
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getAsyncSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, buildAsyncDataSources)
	}
}

func getAsyncSubscriptionData(ctx context.Context, req skill.RequestContext) ([][]edn.RawMessage, skill.Configuration, error) {
	if req.Event.Context.AsyncQueryResult.Name != eventNameAsyncQuery {
		return nil, skill.Configuration{}, nil
	}

	metaEdn, err := b64.StdEncoding.DecodeString(req.Event.Context.AsyncQueryResult.Metadata)
	if err != nil {
		return nil, skill.Configuration{}, fmt.Errorf("failed to decode async metadata: %w", err)
	}

	var metadata data.AsyncResultMetadata
	err = edn.Unmarshal(metaEdn, &metadata)
	if err != nil {
		return nil, skill.Configuration{}, fmt.Errorf("failed to unmarshal async metadata: %w", err)
	}

	return metadata.SubscriptionResults, req.Event.Context.AsyncQueryResult.Configuration, nil
}

// buildAsyncDataSources always returns at least a data.AsyncDataSource,
// but also will return a data.FixedDataSource containing the event payload when applicable
func buildAsyncDataSources(ctx context.Context, req skill.RequestContext) ([]data.DataSource, error) {
	// todo can/should local eval support async queries?
	if req.Event.Context.SyncRequest.Name == eventNameLocalEval {
		return []data.DataSource{}, nil
	}

	if req.Event.Context.AsyncQueryResult.Name != eventNameAsyncQuery {
		return []data.DataSource{
			data.NewAsyncDataSource(req, req.Event.Context.Subscription.Result, map[string]data.AsyncQueryResponse{}),
		}, nil
	}

	metaEdn, err := b64.StdEncoding.DecodeString(req.Event.Context.AsyncQueryResult.Metadata)

	var metadata data.AsyncResultMetadata
	err = edn.Unmarshal(metaEdn, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	var queryResponse data.AsyncQueryResponse
	err = edn.Unmarshal(req.Event.Context.AsyncQueryResult.Result, &queryResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal async query result: %w", err)
	}
	metadata.AsyncQueryResults[metadata.InFlightQueryName] = queryResponse

	return []data.DataSource{
		data.NewAsyncDataSource(req, metadata.SubscriptionResults, metadata.AsyncQueryResults),
	}, nil
}
