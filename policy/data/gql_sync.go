package data

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/hasura/go-graphql-client"

	"github.com/atomist-skills/go-skill"
)

type SyncGraphqlDataSource struct {
	url           string
	graphqlClient *graphql.Client
	logger        skill.Logger
}

func NewSyncGraphqlDataSourceFromSkillRequest(ctx context.Context, req skill.RequestContext) (SyncGraphqlDataSource, error) {
	return NewSyncGraphqlDataSource(ctx, req.Event.Token, req.Event.Urls.Graphql, req.Log)
}

func NewSyncGraphqlDataSource(ctx context.Context, token string, url string, logger skill.Logger) (SyncGraphqlDataSource, error) {
	return NewSyncGraphqlDataSourceWithCorrelationId(ctx, token, url, logger, nil)
}

func NewSyncGraphqlDataSourceWithCorrelationId(ctx context.Context, token string, url string, logger skill.Logger, correlationId *string) (SyncGraphqlDataSource, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token, TokenType: "Bearer"},
	))

	return SyncGraphqlDataSource{
		url: url,
		graphqlClient: graphql.NewClient(url, httpClient).
			WithRequestModifier(func(r *http.Request) {
				r.Header.Add("Accept", "application/json")

				if correlationId != nil {
					r.Header.Add("X-Atomist-Correlation-Id", *correlationId)
				}
			}),
		logger: logger,
	}, nil
}

func (ds SyncGraphqlDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	log := ds.logger

	log.Infof("Graphql endpoint: %s", ds.url)
	log.Infof("Executing query %s: %s", queryName, query)
	log.Infof("Query variables: %v", variables)

	res, err := ds.graphqlClient.ExecRaw(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	log.Infof("GraphQL query response: %s", string(res))

	err = graphql.UnmarshalGraphQL(res, output)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{}, nil
}
