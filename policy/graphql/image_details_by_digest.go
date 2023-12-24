package graphql

import (
	"context"
	"encoding/json"

	"github.com/atomist-skills/go-skill/policy/query"
	"olympos.io/encoding/edn"
)

type ImageDetailsByDigestResponse struct {
	ImageDetailsByDigest *ImageDetailsByDigest `json:"imageDetailsByDigest" edn:"imageDetailsByDigest"`
}

type ImageDetailsByDigest struct {
	Digest       string     `json:"digest" edn:"digest"`
	BaseImage    *BaseImage `json:"baseImage" edn:"baseImage"`
	BaseImageTag *Tag       `json:"baseImageTag" edn:"baseImageTag"`
}

const ImageDetailsAsyncQueryName string = "async-query-image-details"

// GetImageDetailsByDigest fetches and returns a list of base images used in the
// creation of the image specified in digest.
func (client *GraphqlSkillClient) GetImageDetailsByDigest(ctx context.Context, digest string, platform query.ImagePlatform) (*ImageDetailsByDigest, error) {
	log := client.RequestContext.Log

	variables := map[string]interface{}{
		"context": gqlContext(client),
		"digest":  digest,
		"platform": ImagePlatform{
			Architecture: platform.Architecture,
			Os:           platform.Os,
		},
	}

	log.Infof("Executing query: %s with vars %v", baseImagesByDigest, variables)

	res, err := client.GraphqlClient.ExecRaw(ctx, baseImagesByDigest, variables)
	if err != nil {
		return nil, err
	}

	log.Infof("GraphQL query response: %s", string(res))

	var r ImageDetailsByDigestResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return r.ImageDetailsByDigest, nil
}

func (client *GraphqlSkillClient) GetImageDetailsByDigestAsync(ctx context.Context, digest string, platform query.ImagePlatform, asyncClient query.AsyncQueryClient) error {
	log := client.RequestContext.Log

	variables := map[edn.Keyword]interface{}{
		"context": gqlContext(client),
		"digest":  digest,
		"platform": ImagePlatform{
			Architecture: platform.Architecture,
			Os:           platform.Os,
		},
	}

	log.Infof("Executing query: %s with vars %v", baseImagesByDigest, variables)

	log.Infof("Async query")

	query := AsyncGraphqlQuery{
		Query:     baseImagesByDigest,
		Variables: variables,
	}

	return asyncClient.SubmitAsyncQuery(ImageDetailsAsyncQueryName, client.Url, query)
}

func GetImageDetailsByDigestAsyncCallback(response map[edn.Keyword]edn.RawMessage) (*ImageDetailsByDigest, error) {
	result, err := decodeEdnResponse[ImageDetailsByDigestResponse](response)
	if err != nil {
		return nil, err
	}

	if result.ImageDetailsByDigest == nil {
		return nil, nil
	}

	return (*result).ImageDetailsByDigest, nil
}
