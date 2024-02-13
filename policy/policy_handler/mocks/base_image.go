package mocks

import (
	"github.com/atomist-skills/go-skill/policy/types"
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
		FromRepo: nil, // TODO: this data is not present in the SBOM yet
		FromTag:  &sb.Source.Provenance.BaseImage.Tag,
	}
}

func MockSupportedTags(sb *types.SBOM) []string {
	return []string{} // TODO: query GraphQL for supported tags
}
