package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"olympos.io/encoding/edn"

	"github.com/hasura/go-graphql-client"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
)

type SyncGraphqlDataSource struct {
	url           string
	httpClient    http.Client
	logger        skill.Logger
	correlationId *string
	basisT        *int64
}

func NewSyncGraphqlDataSourceFromSkillRequest(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) (SyncGraphqlDataSource, error) {
	return NewSyncGraphqlDataSource(ctx, req.Event.Token, req.Event.Urls.Graphql, req.Log)
}

func NewSyncGraphqlDataSource(ctx context.Context, token string, url string, logger skill.Logger) (SyncGraphqlDataSource, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token, TokenType: "Bearer"},
	))

	return SyncGraphqlDataSource{
		url:        url,
		httpClient: *httpClient,
		logger:     logger,
	}, nil
}

func (ds SyncGraphqlDataSource) WithCorrelationId(correlationId string) SyncGraphqlDataSource {
	return SyncGraphqlDataSource{
		url:           ds.url,
		httpClient:    ds.httpClient,
		correlationId: &correlationId,
		basisT:        ds.basisT,
		logger:        ds.logger,
	}
}

func (ds SyncGraphqlDataSource) WithBasisT(basisT int64) SyncGraphqlDataSource {
	return SyncGraphqlDataSource{
		url:           ds.url,
		httpClient:    ds.httpClient,
		correlationId: ds.correlationId,
		logger:        ds.logger,
		basisT:        &basisT,
	}
}

func (ds SyncGraphqlDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	log := ds.logger

	log.Infof("Graphql endpoint: %s", ds.url)
	log.Infof("Executing query %s: %s", queryName, query)
	log.Infof("Query variables: %v", variables)

	ednVariables := map[edn.Keyword]interface{}{}
	for k, v := range variables {
		ednVariables[edn.Keyword(k)] = v
	}

	res, err := ds.request(ctx, query, ednVariables)
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

func (ds SyncGraphqlDataSource) request(ctx context.Context, query string, variables map[edn.Keyword]interface{}) ([]byte, error) {
	in := GraphQLQueryBody{
		Query:     query,
		Variables: variables,
		BasisT:    ds.basisT,
	}
	var buf bytes.Buffer
	err := edn.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, fmt.Errorf("problem encoding request: %w", err)
	}

	reqReader := bytes.NewReader(buf.Bytes())
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, ds.url, reqReader)
	if err != nil {
		e := fmt.Errorf("problem encoding request: %w", err)

		return nil, e
	}
	request.Header.Add("Content-Type", "application/edn")

	request.Header.Add("Accept", "application/json")

	if ds.correlationId != nil {
		request.Header.Add("X-Atomist-Correlation-Id", *ds.correlationId)
	}

	resp, err := ds.httpClient.Do(request)

	if err != nil {
		e := fmt.Errorf("problem making request: %w", err)
		return nil, e
	}
	defer resp.Body.Close()

	r := resp.Body

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err := fmt.Errorf("%v; body: %q", resp.Status, body)

		return nil, err
	}

	var out struct {
		Data   *json.RawMessage
		Errors graphql.Errors
	}

	err = json.NewDecoder(r).Decode(&out)

	if err != nil {
		return nil, err
	}

	var rawData []byte
	if out.Data != nil && len(*out.Data) > 0 {
		rawData = []byte(*out.Data)
	}

	if len(out.Errors) > 0 {

		return rawData, out.Errors
	}

	return rawData, nil
}
