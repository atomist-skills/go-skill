package mocks

import (
	"context"
	"testing"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/stretchr/testify/assert"
)

type MockDs struct {
	t *testing.T
}

func (ds MockDs) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*data.QueryResponse, error) {
	assert.Equal(ds.t, queryName, imageDetailsByDigestQueryName)
	assert.Equal(ds.t, query, imageDetailsByDigestQuery)

	r := output.(*ImageDetailsByDigestResponse)
	r.ImageDetailsByDigest = &ImageDetailsByDigest{
		Digest: "sha256:1234",
		Repository: Repository{
			HostName: "registry.com",
			RepoName: "namespace/repository",
		},
		Tags: []Tag{
			{
				Name:    "latest",
				Current: true,
			},
			{
				Name:    "1.0",
				Current: false,
			},
		},
	}

	return &data.QueryResponse{}, nil
}

func Test_mockBaseImageDetails_isNotCurrent(t *testing.T) {
	sbom := &types.SBOM{
		Source: types.Source{
			Image: &types.ImageSource{
				Digest: "sha256:9999",
			},
			Provenance: &types.Provenance{
				BaseImage: &types.ProvenanceBaseImage{
					Digest: "sha256:1234",
					Tag:    "1.0",
				},
			},
		},
	}

	logger := skill.Logger{
		Debug:  func(msg string) {},
		Debugf: func(format string, a ...any) {},
	}
	actual, err := mockBaseImageDetails(context.TODO(), skill.RequestContext{Log: logger}, sbom, MockDs{t})
	assert.NoError(t, err)

	expected := ImageDetailsByDigestResponse{
		ImageDetailsByDigest: &ImageDetailsByDigest{
			Digest: "sha256:9999",
			BaseImage: &BaseImage{
				Digest: "sha256:1234",
				Repository: Repository{
					HostName: "registry.com",
					RepoName: "namespace/repository",
				},
				Tags: []Tag{
					{
						Name:    "latest",
						Current: true,
					},
					{
						Name:    "1.0",
						Current: false,
					},
				},
			},
			BaseImageTag: &Tag{
				Name:    "1.0",
				Current: false,
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_mockBaseImageDetails_isCurrent(t *testing.T) {
	sbom := &types.SBOM{
		Source: types.Source{
			Image: &types.ImageSource{
				Digest: "sha256:9999",
			},
			Provenance: &types.Provenance{
				BaseImage: &types.ProvenanceBaseImage{
					Digest: "sha256:1234",
					Tag:    "latest",
				},
			},
		},
	}

	logger := skill.Logger{
		Debug:  func(msg string) {},
		Debugf: func(format string, a ...any) {},
	}
	actual, err := mockBaseImageDetails(context.TODO(), skill.RequestContext{Log: logger}, sbom, MockDs{t})
	assert.NoError(t, err)

	expected := ImageDetailsByDigestResponse{
		ImageDetailsByDigest: &ImageDetailsByDigest{
			Digest: "sha256:9999",
			BaseImage: &BaseImage{
				Digest: "sha256:1234",
				Repository: Repository{
					HostName: "registry.com",
					RepoName: "namespace/repository",
				},
				Tags: []Tag{
					{
						Name:    "latest",
						Current: true,
					},
					{
						Name:    "1.0",
						Current: false,
					},
				},
			},
			BaseImageTag: &Tag{
				Name:    "1.0",
				Current: false,
			},
		},
	}

	assert.Equal(t, expected, actual)
}
