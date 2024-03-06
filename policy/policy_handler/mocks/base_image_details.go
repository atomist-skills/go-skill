package mocks

import (
	"context"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/types"
)

const (
	ImageDetailsQueryName = "image-details-by-digest"

	baseImagesByDigestQueryName = "base-images-by-digest"

	baseImagesByDigestQuery = `
	query ($context: Context!, $digest: String!) {
		baseImagesByDigest(context: $context, digest: $digest) {
		  images {
			digest
			tags {
			  current
			  name
			}
			repository {
				hostName
				repoName
			}
		  }
		}
	}`
)

type (
	GqlImagePlatform struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
		Variant      string `json:"variant,omitempty"`
	}

	Repository struct {
		HostName string `json:"hostName" edn:"hostName"`
		RepoName string `json:"repoName" edn:"repoName"`
	}

	Tag struct {
		Name    string `json:"name" edn:"name"`
		Current bool   `json:"current" edn:"current"`
	}

	BaseImage struct {
		Digest     string     `json:"digest" edn:"digest"`
		Repository Repository `json:"repository" edn:"repository"`
		Tags       []Tag      `json:"tags" edn:"tags"`
	}

	ImageDetailsByDigest struct {
		Digest       string     `json:"digest" edn:"digest"`
		BaseImage    *BaseImage `json:"baseImage" edn:"baseImage"`
		BaseImageTag *Tag       `json:"baseImageTag" edn:"baseImageTag"`
	}

	ImageDetailsByDigestResponse struct {
		ImageDetailsByDigest *ImageDetailsByDigest `json:"imageDetailsByDigest" edn:"imageDetailsByDigest"`
	}

	BaseImagesByDigest struct {
		Images []BaseImage `json:"images" edn:"images"`
	}

	BaseImagesByDigestResponse struct {
		BaseImagesByDigest *BaseImagesByDigest `json:"baseImagesByDigest" edn:"baseImagesByDigest"`
	}
)

func MockBaseImageDetails(ctx context.Context, req skill.RequestContext, sb *types.SBOM) (ImageDetailsByDigestResponse, error) {
	ds, err := data.NewSyncGraphqlDataSource(ctx, req)
	if err != nil {
		return ImageDetailsByDigestResponse{}, err
	}

	return mockBaseImageDetails(ctx, req, sb, ds)
}

func mockBaseImageDetails(ctx context.Context, req skill.RequestContext, sb *types.SBOM, ds data.DataSource) (ImageDetailsByDigestResponse, error) {
	ds, err := data.NewSyncGraphqlDataSource(ctx, req)
	if err != nil {
		return ImageDetailsByDigestResponse{}, err
	}

	baseImageDigest := sb.Source.Provenance.BaseImage.Digest

	var queryResponse BaseImagesByDigestResponse

	queryVariables := map[string]interface{}{"digest": baseImageDigest}
	_, err = ds.Query(ctx, baseImagesByDigestQueryName, baseImagesByDigestQuery, queryVariables, &queryResponse)
	if err != nil {
		return ImageDetailsByDigestResponse{}, err
	}

	if len(queryResponse.BaseImagesByDigest.Images) == 0 {
		return ImageDetailsByDigestResponse{}, fmt.Errorf("no base images found for digest %s", baseImageDigest)
	}

	baseImage := queryResponse.BaseImagesByDigest.Images[0]
	baseImageTag := findBaseImageTag(baseImage, sb.Source.Provenance.BaseImage.Tag)

	mockResponse := ImageDetailsByDigestResponse{
		ImageDetailsByDigest: &ImageDetailsByDigest{
			Digest:       sb.Source.Image.Digest,
			BaseImage:    &baseImage,
			BaseImageTag: baseImageTag,
		},
	}

	return mockResponse, nil
}

func findBaseImageTag(baseImage BaseImage, tag string) *Tag {
	for _, t := range baseImage.Tags {
		if t.Name == tag {
			return &t
		}
	}
	return nil
}
