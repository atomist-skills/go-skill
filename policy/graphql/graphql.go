package graphql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atomist-skills/go-skill/policy/query"

	"github.com/atomist-skills/go-skill"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"olympos.io/encoding/edn"
)

type GraphqlSkillClient struct {
	Url              string
	GraphqlClient    *graphql.Client
	AsyncQueryClient query.AsyncQueryClient
	RequestContext   skill.RequestContext
}

type AsyncGraphqlQuery struct {
	Query     string                      `edn:"query"`
	Variables map[edn.Keyword]interface{} `edn:"variables"`
}

type GraphqlClient interface {
	GetVulnerabilitiesByPackage(ctx context.Context, purls []string) ([]VulnerabilitiesByPackage, error)
	GetImageDetailsByDigest(ctx context.Context, digest string, platform query.ImagePlatform) (*ImageDetailsByDigest, error)
	GetImagePackagesByDigest(ctx context.Context, digest string) (*ImagePackagesByDigest, error)

	GetImagePackagesByDigestAsync(ctx context.Context, digest string, asyncClient query.AsyncQueryClient) error
	GetImageDetailsByDigestAsync(ctx context.Context, digest string, platform query.ImagePlatform, asyncClient query.AsyncQueryClient) error

	GetRequestContext() skill.RequestContext
}

func (client *GraphqlSkillClient) GetRequestContext() skill.RequestContext {
	return client.RequestContext
}

func NewGraphqlSkillClient(ctx context.Context, req skill.RequestContext) (GraphqlClient, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: req.Event.Token, TokenType: "Bearer"},
	))

	return &GraphqlSkillClient{
		Url: req.Event.Urls.Graphql,
		GraphqlClient: graphql.NewClient(req.Event.Urls.Graphql, httpClient).
			WithRequestModifier(func(r *http.Request) {
				r.Header.Add("Accept", "application/json")
			}),
		RequestContext: req,
	}, nil
}

func gqlContext(client *GraphqlSkillClient) GqlContext {
	return GqlContext{
		TeamId:       client.RequestContext.Event.WorkspaceId,
		Organization: client.RequestContext.Event.Organization,
	}
}

func decodeEdnResponse[P interface{}](result map[edn.Keyword]edn.RawMessage) (*P, error) {
	type resultType struct {
		Data   P `edn:"data"`
		Errors []struct {
			Message string `edn:"message"`
		} `edn:"errors"`
	}

	ednboby, err := edn.Marshal(result)
	if err != nil {
		return nil, err
	}
	var decoded resultType
	err = edn.Unmarshal(ednboby, &decoded)
	if err != nil {
		return nil, err
	}

	if len(decoded.Errors) > 0 {
		return nil, fmt.Errorf(decoded.Errors[0].Message)
	}
	return &decoded.Data, nil
}
