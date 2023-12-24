package policy_handler

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"olympos.io/encoding/edn"
)

const eventNameAsyncQuery = "async_query"

// WithAsync will enable async graphql queries for the EventHandler.
// When used, data.QueryResponse#AsyncRequestMade will be true when performed asynchronously.
// If that flag is set, the policy evaluator should terminate early with no results.
// It will be automatically retried once the async query results are returned.
func WithAsync() Opt {
	return func(h *EventHandler) error {
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getAsyncSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, buildAsyncDataSources)
		return nil
	}
}

func getAsyncSubscriptionData(ctx context.Context, req skill.RequestContext) ([][]edn.RawMessage, error) {
	metadata := req.Event.Context.AsyncQueryResult.Metadata
	encoded, err := b64.StdEncoding.DecodeString(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode async metadata: %w", err)
	}

	var subscriptionResult [][]edn.RawMessage
	err = edn.Unmarshal(encoded, &subscriptionResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal async metadata: %w", err)
	}

	return subscriptionResult, nil
}

// buildAsyncDataSources always returns at least a data.AsyncDataSource,
// but also will return a data.FixedDataSource containing the event payload when applicable
func buildAsyncDataSources(ctx context.Context, req skill.RequestContext) ([]data.DataSource, error) {
	if req.Event.Context.AsyncQueryResult.Name == eventNameAsyncQuery {
		responseSource, err := data.UnwrapAsyncResponse(req.Event.Context.AsyncQueryResult.Result)
		if err != nil {
			return nil, err
		}
		return []data.DataSource{
			responseSource,
			data.NewAsyncDataSource(req, req.Event.Context.AsyncQueryResult.Metadata),
		}, nil
	}

	edn, err := edn.Marshal(req.Event.Context.Subscription.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata [%w]", err)
	}

	return []data.DataSource{
		data.NewAsyncDataSource(req, b64.StdEncoding.EncodeToString(edn)),
	}, nil
}
