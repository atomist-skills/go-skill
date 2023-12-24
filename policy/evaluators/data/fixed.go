package data

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/graphql"
	"github.com/atomist-skills/go-skill/policy/query"
)

// FixedDataSource is only used for tests
type FixedDataSource struct {
	Packages             map[string][]Package
	ImageDetailsByDigest *graphql.ImageDetailsByDigest
}

func (s FixedDataSource) GetPackages(ctx context.Context, digest string) (*GetPackagesResult, error) {
	return &GetPackagesResult{
		AsyncQueryMade: false,
		Result:         s.Packages[digest],
	}, nil
}

func (s FixedDataSource) GetImageDetailsByDigest(ctx context.Context, digest string, platform query.ImagePlatform) (*GetImageDetailsByDigestResult, error) {
	return &GetImageDetailsByDigestResult{
		AsyncQueryMade: false,
		Result:         s.ImageDetailsByDigest,
	}, nil
}
