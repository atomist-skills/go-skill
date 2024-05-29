package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/hasura/go-graphql-client"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data/cache"
	"github.com/atomist-skills/go-skill/policy/goals"
)

type SyncGraphqlQueryClient struct {
	url           string
	httpClient    http.Client
	logger        skill.Logger
	correlationId *string
	basisT        *int64
	cache         *cache.QueryCache
	retryBackoff  time.Duration
}

type SyncGraphQLQueryBody struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	BasisT    *int64                 `json:"basis-t,omitempty"`
}

func NewSyncGraphqlQueryClientFromSkillRequest(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) SyncGraphqlQueryClient {
	return NewSyncGraphqlQueryClient(ctx, req.Event.Token, req.Event.Urls.Graphql, req.Log).WithBasisT(evalMeta.SubscriptionBasisT)
}

func NewSyncGraphqlQueryClient(ctx context.Context, token string, url string, logger skill.Logger) SyncGraphqlQueryClient {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token, TokenType: "Bearer"},
	))

	return SyncGraphqlQueryClient{
		url:          url,
		httpClient:   *httpClient,
		logger:       logger,
		retryBackoff: 10 * time.Second,
	}
}

func (ds SyncGraphqlQueryClient) WithCorrelationId(correlationId string) SyncGraphqlQueryClient {
	ds.correlationId = &correlationId

	return ds
}

func (ds SyncGraphqlQueryClient) WithBasisT(basisT int64) SyncGraphqlQueryClient {
	if basisT == 0 {
		ds.basisT = nil
	} else {
		ds.basisT = &basisT
	}

	return ds
}

func (ds SyncGraphqlQueryClient) WithQueryCache(cache cache.QueryCache) SyncGraphqlQueryClient {
	ds.cache = &cache

	return ds
}

func (ds SyncGraphqlQueryClient) WithRetryBackoff(backoff time.Duration) SyncGraphqlQueryClient {
	ds.retryBackoff = backoff

	return ds
}

func (ds SyncGraphqlQueryClient) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	log := ds.logger

	log.Infof("Graphql endpoint: %s", ds.url)
	log.Infof("Executing query %s: %s", queryName, query)
	log.Debugf("Query variables: %v", variables)

	res, err := ds.requestWithCache(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	err = graphql.UnmarshalGraphQL(res, output)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{}, nil
}

func (ds SyncGraphqlQueryClient) requestWithCache(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	if ds.cache != nil {
		res, err := (*ds.cache).Read(ctx, query, variables)
		if err != nil {
			return nil, err
		}

		if res != nil {
			ds.logger.Info("Cache hit for query")
			return res, nil
		}
	}

	res, err := ds.request(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	if ds.cache != nil {
		err = (*ds.cache).Write(ctx, query, variables, res)
	}

	return res, err
}

func (ds SyncGraphqlQueryClient) request(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	in := SyncGraphQLQueryBody{
		Query:     query,
		Variables: variables,
		BasisT:    ds.basisT,
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, fmt.Errorf("problem encoding request: %w", err)
	}

	reqReader := bytes.NewReader(buf.Bytes())
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, ds.url, reqReader)
	if err != nil {
		e := fmt.Errorf("problem encoding request: %w", err)

		return nil, e
	}
	request.Header.Add("Content-Type", "application/json")

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

	if resp.StatusCode >= 500 && ds.retryBackoff > 0 {
		time.Sleep(ds.retryBackoff)

		resp, err = ds.httpClient.Do(request)
		if err != nil {
			e := fmt.Errorf("problem making request: %w", err)
			return nil, e
		}
		defer resp.Body.Close()
	}

	r := resp.Body

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
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

	ds.logger.Debugf("Sync GQL query response: %s", string(rawData))

	return rawData, nil
}
