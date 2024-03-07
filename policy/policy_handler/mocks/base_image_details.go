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

	imageDetailsByDigestQueryName = "base-image-details-by-digest"

	imageDetailsByDigestQuery = `
	query ($context: Context!, $digest: String!, $platform: ImagePlatform!) {
		imageDetailsByDigest(context: $context, digest: $digest, platform: $platform) {
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
		Tags         []Tag      `json:"tags" edn:"tags"`
		Repository   Repository `json:"repository" edn:"repository"`
	}

	ImageDetailsByDigestResponse struct {
		ImageDetailsByDigest *ImageDetailsByDigest `json:"imageDetailsByDigest" edn:"imageDetailsByDigest"`
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
	baseImageDigest := sb.Source.Provenance.BaseImage.Digest

	var queryResponse ImageDetailsByDigestResponse

	queryVariables := map[string]interface{}{
		"digest":  baseImageDigest,
		"context": data.GqlContext(req),
		"platform": GqlImagePlatform{
			Architecture: sb.Source.Provenance.BaseImage.Platform.Architecture,
			Os:           sb.Source.Provenance.BaseImage.Platform.OS,
			Variant:      sb.Source.Provenance.BaseImage.Platform.Variant}}
	_, err := ds.Query(ctx, imageDetailsByDigestQueryName, imageDetailsByDigestQuery, queryVariables, &queryResponse)
	if err != nil {
		return ImageDetailsByDigestResponse{}, err
	}

	if queryResponse.ImageDetailsByDigest == nil {
		return ImageDetailsByDigestResponse{}, fmt.Errorf("no base images found for digest %s", baseImageDigest)
	}

	baseImage := queryResponse.ImageDetailsByDigest
	baseImageTag := findBaseImageTag(*baseImage, sb.Source.Provenance.BaseImage.Tag)

	mockResponse := ImageDetailsByDigestResponse{
		ImageDetailsByDigest: &ImageDetailsByDigest{
			Digest: sb.Source.Image.Digest,
			BaseImage: &BaseImage{
				Digest:     baseImage.Digest,
				Repository: baseImage.Repository,
				Tags:       baseImage.Tags,
			},
			BaseImageTag: baseImageTag,
		},
	}

	return mockResponse, nil
}

func findBaseImageTag(baseImage ImageDetailsByDigest, tag string) *Tag {
	for _, t := range baseImage.Tags {
		if t.Name == tag {
			return &t
		}
	}
	return nil
}
