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
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"olympos.io/encoding/edn"
)

const eventNameLocalEval = "evaluate_goals_locally"

type SyncRequestMetadata struct {
	QueryResults map[edn.Keyword]edn.RawMessage `edn:"fixedQueryResults"`
	Packages     []Package                      `edn:"packages"`      // todo remove when no longer used
	User         string                         `edn:"imgConfigUser"` // The user from the image config blob // todo remove when no longer used
	SBOM         string                         `edn:"sbom"`
	ContentType  string                         `edn:"contentType"`
	Encoding     string                         `edn:"encoding"`
}

type Package struct {
	Licenses  []string `edn:"licenses,omitempty"` // only needed for the license policy evaluation
	Name      string   `edn:"name"`
	Namespace string   `edn:"namespace"`
	Version   string   `edn:"version"`
	Purl      string   `edn:"purl"`
	Type      string   `edn:"type"`
}

func WithLocal() Opt {
	return func(h *EventHandler) {
		h.subscriptionNames = append(h.subscriptionNames, eventNameLocalEval)
		h.evalInputProviders = append(h.evalInputProviders, getLocalSubscriptionData)
		h.transactFilters = append(h.transactFilters, shouldTransactLocal)
	}
}

func getLocalSubscriptionData(_ context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, *types.SBOM, error) {
	if req.Event.Context.SyncRequest.Name != eventNameLocalEval {
		return nil, skill.Configuration{}, nil, nil
	}

	syncRequestMetadata, sbom, err := parseMetadata(req)
	if err != nil {
		return nil, skill.Configuration{}, nil, err
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

		artifacts := []types.Package{}
		for _, pkg := range syncRequestMetadata.Packages {
			artifacts = append(artifacts, types.Package{
				Name:      pkg.Name,
				Version:   pkg.Version,
				Type:      pkg.Type,
				Purl:      pkg.Purl,
				Licenses:  pkg.Licenses,
				Namespace: pkg.Namespace,
			})
		}

		sbom = &types.SBOM{
			Source: types.Source{
				Image: &types.ImageSource{
					Digest: "localDigest",
					Config: &v1.ConfigFile{
						Config: v1.Config{
							User: syncRequestMetadata.User,
						},
					},
				},
			},
			Artifacts: artifacts,
		}
	}

	subscriptionData, err := edn.Marshal(mockCommonSubscriptionData)
	if err != nil {
		return nil, skill.Configuration{}, nil, err
	}

	subscriptionResult := map[edn.Keyword]edn.RawMessage{}
	subscriptionResult[edn.Keyword("image")] = subscriptionData

	return &goals.EvaluationMetadata{
		SubscriptionResult: []map[edn.Keyword]edn.RawMessage{
			subscriptionResult,
		}}, req.Event.Context.SyncRequest.Configuration, sbom, nil
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
