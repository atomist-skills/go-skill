package policy_handler

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"olympos.io/encoding/edn"
)

const eventNameAsyncQuery = data.AsyncQueryName // these must match for the event handler to be registered

// WithAsyncMultiQuery will enable the async graphql data source to spool results across multiple queries.
// These intermediate results are stored in the following requests' metadata,
// and as such risk hitting the upper limit on the metadata field, and failing.
func WithAsyncMultiQuery() Opt {
	return func(h *EventHandler) {
		h.subscriptionNames = append(h.subscriptionNames, eventNameAsyncQuery)
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getAsyncSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, buildAsyncDataSources(true))
	}
}

// withAsync is enabled by default, added last after all other Opts.
func withAsync() Opt {
	return func(h *EventHandler) {
		// don't register if WithAsyncMultiQuery is already enabled
		for _, s := range h.subscriptionNames {
			if s == eventNameAsyncQuery {
				return
			}
		}

		h.subscriptionNames = append(h.subscriptionNames, eventNameAsyncQuery)
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getAsyncSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, buildAsyncDataSources(false))
	}
}

func getAsyncSubscriptionData(ctx context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, error) {
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

	return &metadata.EvaluationMetadata, req.Event.Context.AsyncQueryResult.Configuration, nil
}

func buildAsyncDataSources(multipleQuerySupport bool) dataSourceProvider {
	return func(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
		if req.Event.Context.SyncRequest.Name == eventNameLocalEval {
			return []data.DataSource{}, nil
		}

		if req.Event.Context.AsyncQueryResult.Name != eventNameAsyncQuery {
			return []data.DataSource{
				data.NewAsyncDataSource(multipleQuerySupport, req, evalMeta, map[string]data.AsyncQueryResponse{}),
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
			data.NewAsyncDataSource(multipleQuerySupport, req, metadata.EvaluationMetadata, metadata.AsyncQueryResults),
		}, nil
	}
}
