package policy_handler

import (
	"context"
	"fmt"
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/policy_handler/legacy"
	"olympos.io/encoding/edn"
)

const eventNameLocalEval = "evaluate_goals_locally"

type SyncRequestMetadata struct {
	QueryResults map[edn.Keyword]edn.RawMessage `edn:"fixedQueryResults"`
	Packages     []legacy.Package               `edn:"packages"`      // todo remove when no longer used
	User         string                         `edn:"imgConfigUser"` // The user from the image config blob // todo remove when no longer used
}

func WithLocal() Opt {
	return func(h *EventHandler) {
		h.subscriptionNames = append(h.subscriptionNames, eventNameLocalEval)
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getLocalSubscriptionData)
		h.dataSourceProviders = append(h.dataSourceProviders, buildLocalDataSources)
		h.transactFilters = append(h.transactFilters, shouldTransactLocal)
	}
}

func getLocalSubscriptionData(ctx context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return nil, skill.Configuration{}, nil
	}

	mockCommonSubscriptionData := goals.CommonSubscriptionQueryResult{
		ImageDigest: "localDigest",
	}
	subscriptionData, err := edn.Marshal(mockCommonSubscriptionData)
	if err != nil {
		return nil, skill.Configuration{}, err
	}

	return &goals.EvaluationMetadata{
		SubscriptionResult: [][]edn.RawMessage{{subscriptionData}},
	}, req.Event.Context.SyncRequest.Configuration, nil
}

func buildLocalDataSources(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return []data.DataSource{}, nil
	}

	var srMeta SyncRequestMetadata
	err := edn.Unmarshal(req.Event.Context.SyncRequest.Metadata, &srMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SyncRequest metadata: %w", err)
	}

	fixedQueryResults := map[string][]byte{}
	for k, v := range srMeta.QueryResults {
		fixedQueryResults[string(k)] = v
	}

	if _, ok := fixedQueryResults[legacy.ImagePackagesByDigestQueryName]; !ok && len(srMeta.Packages) != 0 {
		mockedQueryResult, err := legacy.MockImagePackagesByDigest(ctx, req, srMeta.Packages)
		if err != nil {
			return nil, err
		}

		pkgsEdn, err := edn.Marshal(mockedQueryResult)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal mocked query %s: %w", legacy.ImagePackagesByDigestQueryName, err)
		}

		fixedQueryResults[legacy.ImagePackagesByDigestQueryName] = pkgsEdn
	}

	return []data.DataSource{
		data.NewFixedDataSource(edn.Unmarshal, fixedQueryResults),
	}, nil
}

func shouldTransactLocal(ctx context.Context, req skill.RequestContext) bool {
	return req.Event.Context.SyncRequest.Name != eventNameLocalEval
}
