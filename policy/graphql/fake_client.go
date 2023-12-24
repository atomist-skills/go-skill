package graphql

import (
	"context"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/internal/test_util"
	"github.com/atomist-skills/go-skill/policy/query"
)

type FakeGraphqlClient struct {
	ImageDetailsByDigest  *ImageDetailsByDigest
	ImagePackagesByDigest ImagePackagesByDigest
}

func (client *FakeGraphqlClient) GetImageDetailsByDigest(ctx context.Context, digest string, platform query.ImagePlatform) (*ImageDetailsByDigest, error) {
	return client.ImageDetailsByDigest, nil
}

func (client *FakeGraphqlClient) GetImagePackagesByDigest(ctx context.Context, digest string) (ImagePackagesByDigest, error) {
	return client.ImagePackagesByDigest, nil
}

func (client *FakeGraphqlClient) GetRequestContext() skill.RequestContext {
	return skill.RequestContext{
		Log: test_util.CreateEmptyLogger(),
	}
}
