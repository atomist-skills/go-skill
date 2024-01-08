package data

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/atomist-skills/go-skill"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

type SyncGraphqlDataSource struct {
	Url            string
	GraphqlClient  *graphql.Client
	RequestContext skill.RequestContext
}

func NewSyncGraphqlDataSource(ctx context.Context, req skill.RequestContext) (SyncGraphqlDataSource, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: req.Event.Token, TokenType: "Bearer"},
	))

	return SyncGraphqlDataSource{
		Url: req.Event.Urls.Graphql,
		GraphqlClient: graphql.NewClient(req.Event.Urls.Graphql, httpClient).
			WithRequestModifier(func(r *http.Request) {
				r.Header.Add("Accept", "application/json")
			}),
		RequestContext: req,
	}, nil
}

func (ds SyncGraphqlDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	log := ds.RequestContext.Log

	log.Infof("Graphql endpoint: %s", ds.RequestContext.Event.Urls.Graphql)
	log.Infof("Executing query %s: %s", queryName, query)
	log.Infof("Query variables: %v", variables)

	res, err := ds.GraphqlClient.ExecRaw(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	log.Infof("GraphQL query response: %s", string(res))

	err = json.Unmarshal(res, output)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{}, nil
}
