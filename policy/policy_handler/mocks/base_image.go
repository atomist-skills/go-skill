package mocks

import (
	"fmt"
	"github.com/atomist-skills/go-skill/policy/types"
	"strings"
)

const (
	GetBaseImageQueryName     = "get-base-image"
	GetSupportedTagsQueryName = "get-supported-tags"
)

type BaseImageQueryResult struct {
	FromReference *SubscriptionImage      `edn:"docker.image/from"`
	FromRepo      *SubscriptionRepository `edn:"docker.image/from-repository"`
	FromTag       *string                 `edn:"docker.image/from-tag"`
}

type SubscriptionImage struct {
	Digest string `edn:"docker.image/digest"`
}

type SubscriptionRepository struct {
	Host       string `edn:"docker.repository/host"`
	Repository string `edn:"docker.repository/repository"`
}

func MockBaseImage(sb *types.SBOM) BaseImageQueryResult {
	return BaseImageQueryResult{
		FromReference: &SubscriptionImage{
			Digest: sb.Source.Provenance.BaseImage.Digest,
		},
		FromRepo: parseFromReference(sb),
		FromTag:  &sb.Source.Provenance.BaseImage.Tag,
	}
}

func parseFromReference(sb *types.SBOM) *SubscriptionRepository {
	// this is registry.com/namespace/repository form
	// but minified (omits hub.docker.com and library/ if unnecessary)
	fullName := sb.Source.Provenance.BaseImage.Name
	if fullName == "" {
		return nil
	}

	parts := strings.SplitN(fullName, "/", 3)
	switch len(parts) {
	case 1:
		return &SubscriptionRepository{
			Host:       "hub.docker.com",
			Repository: fmt.Sprintf("library/%s", parts[0]),
		}

	case 2:
		return &SubscriptionRepository{
			Host:       "hub.docker.com",
			Repository: fmt.Sprintf("%s/%s", parts[0], parts[1]),
		}

	default:
		return &SubscriptionRepository{
			Host:       parts[0],
			Repository: fmt.Sprintf("%s/%s", parts[1], parts[2]),
		}
	}
}

func MockSupportedTags(sb *types.SBOM) []string {
	return []string{} // TODO: query GraphQL for supported tags
}
