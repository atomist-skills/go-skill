package graphql

import (
	"context"
	"encoding/json"

	"github.com/atomist-skills/go-skill/policy/query"

	"olympos.io/encoding/edn"
)

type ImagePackagesByDigestResponse struct {
	ImagePackagesByDigest *ImagePackagesByDigest `json:"imagePackagesByDigest" edn:"imagePackagesByDigest"`
}

type ImagePackagesByDigest struct {
	Digest         string         `json:"digest" edn:"digest"`
	ImagePackages  ImagePackages  `json:"imagePackages" edn:"imagePackages"`
	ImageHistories []ImageHistory `json:"imageHistories" edn:"imageHistories"`
	ImageLayers    ImageLayers    `json:"imageLayers" edn:"imageLayers"`
}

const ImagePackagesAsyncQueryName string = "async-query-packages"

func (client *GraphqlSkillClient) GetImagePackagesByDigest(ctx context.Context, digest string) (*ImagePackagesByDigest, error) {
	log := client.RequestContext.Log

	variables := map[string]interface{}{
		"context": gqlContext(client),
		"digest":  digest,
	}

	log.Infof("Graphql endpoint: %s", client.RequestContext.Event.Urls.Graphql)
	log.Infof("Executing query: %s", imagePackagesByDigestQuery)
	log.Infof("Query variables: %v", variables)

	res, err := client.GraphqlClient.ExecRaw(ctx, imagePackagesByDigestQuery, variables)
	if err != nil {
		return nil, err
	}

	log.Infof("GraphQL query response: %s", string(res))

	imagePackages, err := getPackagesFromJsonResponse(res)
	if err != nil {
		return nil, err
	}

	if imagePackages != nil {
		log.Infof("Found %d packages for digest %s", len(imagePackages.ImagePackages.Packages), digest)
	} else {
		log.Infof("Empty package response for digest %s", digest)
	}

	return imagePackages, nil
}

func (client *GraphqlSkillClient) GetImagePackagesByDigestAsync(ctx context.Context, digest string, asyncClient query.AsyncQueryClient) error {
	log := client.RequestContext.Log

	variables := map[edn.Keyword]interface{}{
		"context": gqlContext(client),
		"digest":  digest,
	}

	log.Infof("Graphql endpoint: %s", client.RequestContext.Event.Urls.Graphql)
	log.Infof("Executing query: %s", imagePackagesByDigestQuery)
	log.Infof("Query variables: %v", variables)
	log.Infof("Async query")

	query := AsyncGraphqlQuery{
		Query:     imagePackagesByDigestQuery,
		Variables: variables,
	}

	return asyncClient.SubmitAsyncQuery(ImagePackagesAsyncQueryName, client.Url, query)
}

func GetImagePackagesByDigestAsyncCallback(response map[edn.Keyword]edn.RawMessage) (*ImagePackagesByDigest, error) {
	result, err := decodeEdnResponse[ImagePackagesByDigestResponse](response)
	if err != nil {
		return nil, err
	}

	return (*result).ImagePackagesByDigest, nil
}

func getPackagesFromJsonResponse(b []byte) (*ImagePackagesByDigest, error) {
	var r ImagePackagesByDigestResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}

	return r.ImagePackagesByDigest, nil
}
