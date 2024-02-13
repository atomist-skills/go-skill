package mocks

import (
	"context"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/policy_handler/legacy"
	"github.com/atomist-skills/go-skill/policy/types"
	"olympos.io/encoding/edn"
)

func BuildLocalEvalMocks(ctx context.Context, req skill.RequestContext, sb *types.SBOM) (map[edn.Keyword]edn.RawMessage, error) {
	m := map[edn.Keyword]edn.RawMessage{}
	if sb == nil {
		req.Log.Info("No SBOM provided, returning empty map")
		return m, nil
	}

	// Image packages by digest
	imgPkgsMock, err := MockImagePackagesByDigest(ctx, req, sb)
	if err != nil {
		return m, err
	}
	m[legacy.ImagePackagesByDigestQueryName], err = edn.Marshal(imgPkgsMock)
	if err != nil {
		return m, fmt.Errorf("failed to marshal image packages by digest mock: %w", err)
	}

	// User
	if sb.Source.Image != nil && sb.Source.Image.Config != nil {
		userMock := MockGetUser(sb.Source.Image.Config.Config.User)
		m[GetUserQueryName], err = edn.Marshal(userMock)
		if err != nil {
			return m, fmt.Errorf("failed to marshal get user mock: %w", err)
		}
	}

	// Attestations
	req.Log.Infof("SBOM has %d attestations", len(sb.Attestations))
	if len(sb.Attestations) > 0 {
		attestMock := MockGetInTotoAttestations(sb, req.Log)
		m[GetInTotoAttestationsQueryName], err = edn.Marshal(attestMock)
		if err != nil {
			return m, fmt.Errorf("failed to marshal attestations mock: %w", err)
		}
	}

	// Base image
	if sb.Source.Provenance == nil {
		req.Log.Info("Skipping base image mock, no provenance in SBOM")
	} else {
		baseImageMock := MockBaseImage(sb)
		m[GetBaseImageQueryName], err = edn.Marshal(baseImageMock)
		if err != nil {
			return m, fmt.Errorf("failed to marshal base image mock: %w", err)
		}
	}

	// Supported tags
	supportedTagsMock := MockSupportedTags(sb)
	m[GetSupportedTagsQueryName], err = edn.Marshal(supportedTagsMock)
	if err != nil {
		return m, fmt.Errorf("failed to marshal supported tags mock: %w", err)
	}

	// TODO: Mock ImagePlatforms from CommonSubscriptionQueryResult (required for atomist/no-stale-base-images)
	// TODO: Mock packages (required for atomist/license-goal)

	return m, nil
}
