package policy_handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/policy_handler/legacy"
	"github.com/atomist-skills/go-skill/policy/policy_handler/mocks"
	"github.com/atomist-skills/go-skill/policy/types"
	"olympos.io/encoding/edn"
)

const eventNameLocalEval = "evaluate_goals_locally"

type SyncRequestMetadata struct {
	QueryResults map[edn.Keyword]edn.RawMessage `edn:"fixedQueryResults"`
	Packages     []legacy.Package               `edn:"packages"`      // todo remove when no longer used
	User         string                         `edn:"imgConfigUser"` // The user from the image config blob // todo remove when no longer used
	SBOM         string                         `edn:"sbom"`
	ContentType  string                         `edn:"contentType"`
	Encoding     string                         `edn:"encoding"`
}

func WithLocal() Opt {
	return func(h *EventHandler) {
		h.subscriptionNames = append(h.subscriptionNames, eventNameLocalEval)
		h.subscriptionDataProviders = append(h.subscriptionDataProviders, getLocalSubscriptionData)
		h.dataSourceProviders = append([]dataSourceProvider{buildLocalDataSources}, h.dataSourceProviders...)
		h.transactFilters = append(h.transactFilters, shouldTransactLocal)
	}
}

func getLocalSubscriptionData(_ context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return nil, skill.Configuration{}, nil
	}

	_, sbom, err := parseMetadata(req)
	if err != nil {
		return nil, skill.Configuration{}, err
	}

	var mockCommonSubscriptionData goals.ImageSubscriptionQueryResult
	if sbom != nil {
		mockCommonSubscriptionData = goals.ImageSubscriptionQueryResult{
			ImageDigest: sbom.Source.Image.Digest,
			ImagePlatforms: []goals.ImagePlatform{{
				Architecture: sbom.Source.Image.Platform.Architecture,
				Os:           sbom.Source.Image.Platform.Os,
			}},
		}
	} else {
		mockCommonSubscriptionData = goals.ImageSubscriptionQueryResult{
			ImageDigest: "localDigest",
		}
	}

	subscriptionData, err := edn.Marshal(mockCommonSubscriptionData)
	if err != nil {
		return nil, skill.Configuration{}, err
	}

	return &goals.EvaluationMetadata{
		SubscriptionResult: [][]edn.RawMessage{{subscriptionData}},
	}, req.Event.Context.SyncRequest.Configuration, nil
}

func buildLocalDataSources(ctx context.Context, req skill.RequestContext, _ goals.EvaluationMetadata) ([]data.DataSource, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return []data.DataSource{}, nil
	}

	srMeta, sbom, err := parseMetadata(req)
	if err != nil {
		return nil, err
	}

	srMeta.QueryResults, err = mocks.BuildLocalEvalMocks(ctx, req, sbom)
	if err != nil {
		return nil, fmt.Errorf("failed to build local evaluation mocks: %w", err)
	}

	fixedQueryResults := map[string][]byte{}
	for k, v := range srMeta.QueryResults {
		fixedQueryResults[string(k)] = v
	}

	if _, ok := fixedQueryResults[legacy.ImagePackagesByDigestQueryName]; !ok && len(srMeta.Packages) != 0 {
		mockedQueryResult, err := legacy.MockImagePackagesByDigest(ctx, req, srMeta.Packages, nil)
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

func shouldTransactLocal(_ context.Context, req skill.RequestContext) bool {
	return req.Event.Context.SyncRequest.Name != eventNameLocalEval
}

func parseMetadata(req skill.RequestContext) (SyncRequestMetadata, *types.SBOM, error) {
	var srMeta SyncRequestMetadata
	err := edn.Unmarshal(req.Event.Context.SyncRequest.Metadata, &srMeta)
	if err != nil {
		return SyncRequestMetadata{}, nil, fmt.Errorf("failed to unmarshal SyncRequest metadata: %w", err)
	}

	if srMeta.SBOM == "" {
		return srMeta, nil, nil
	}

	decodedSBOM, err := base64.StdEncoding.DecodeString(srMeta.SBOM)
	if err != nil {
		return srMeta, nil, fmt.Errorf("failed to base64-decode SBOM: %w", err)
	}
	if srMeta.Encoding == "base64+gzip" {
		reader := bytes.NewReader(decodedSBOM)
		gzreader, err := gzip.NewReader(reader)
		defer gzreader.Close() //nolint:errcheck
		if err != nil {
			return srMeta, nil, fmt.Errorf("failed to decompress SBOM: %w", err)
		}
		decodedSBOM, err = io.ReadAll(gzreader)
		if err != nil {
			return srMeta, nil, fmt.Errorf("failed to base64-decode SBOM: %w", err)
		}
	}

	var sbom *types.SBOM
	// THE SBOM is a JSON here, not edn
	if err := json.Unmarshal(decodedSBOM, &sbom); err != nil {
		return srMeta, nil, fmt.Errorf("failed to unmarshal SBOM: %w", err)
	}

	return srMeta, sbom, nil
}
