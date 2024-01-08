package data

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
	"olympos.io/encoding/edn"
)

const AsyncQueryName = "async-query"

type (
	AsyncQueryBody struct {
		Query     string                      `edn:"query"`
		Variables map[edn.Keyword]interface{} `edn:"variables"`
	}

	AsyncQueryRequest struct {
		Name     string         `edn:"name"`
		Body     AsyncQueryBody `edn:"body"`
		Metadata string         `edn:"metadata"`
	}

	AsyncQueryResponse struct {
		Data   edn.RawMessage `edn:"data"`
		Errors []struct {
			Message string `edn:"message"`
		}
	}

	AsyncResultMetadata struct {
		EvaluationMetadata goals.EvaluationMetadata      `edn:"evalMeta"`
		AsyncQueryResults  map[string]AsyncQueryResponse `edn:"results"`
		InFlightQueryName  string                        `edn:"query-name"`
	}

	AsyncDataSource struct {
		multipleQuerySupport bool
		log                  skill.Logger
		url                  string
		token                string
		evaluationMetadata   goals.EvaluationMetadata
		asyncResults         map[string]AsyncQueryResponse
	}
)

func NewAsyncDataSource(
	multipleQuerySupport bool,
	req skill.RequestContext,
	evaluationMetadata goals.EvaluationMetadata,
	asyncResults map[string]AsyncQueryResponse,
) AsyncDataSource {
	return AsyncDataSource{
		multipleQuerySupport: multipleQuerySupport,
		log:                  req.Log,
		url:                  fmt.Sprintf("%s:enqueue", req.Event.Urls.Graphql),
		token:                req.Event.Token,
		evaluationMetadata:   evaluationMetadata,
		asyncResults:         asyncResults,
	}
}

func (ds AsyncDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	if existingResult, ok := ds.asyncResults[queryName]; ok {
		if len(existingResult.Errors) != 0 {
			return nil, fmt.Errorf("async query returned error: %s", existingResult.Errors[0].Message)
		}
		return &QueryResponse{}, edn.Unmarshal(existingResult.Data, output)
	}

	if len(ds.asyncResults) > 0 && !ds.multipleQuerySupport {
		ds.log.Debugf("skipping async query for query %s due to lack of multipleQuerySupport", queryName)
		return nil, nil // don't error, in case there is another applicable query executor down-chain
	}

	ednVariables := map[edn.Keyword]interface{}{}
	for k, v := range variables {
		ednVariables[edn.Keyword(k)] = v
	}

	metadata := AsyncResultMetadata{
		EvaluationMetadata: ds.evaluationMetadata,
		AsyncQueryResults:  ds.asyncResults,
		InFlightQueryName:  queryName,
	}
	metadataEdn, err := edn.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	request := AsyncQueryRequest{
		Name: AsyncQueryName,
		Body: AsyncQueryBody{
			Query:     query,
			Variables: ednVariables,
		},
		Metadata: b64.StdEncoding.EncodeToString(metadataEdn),
	}

	reqEdn, err := edn.Marshal(request)
	if err != nil {
		return nil, err
	}

	ds.log.Infof("Async request: %s", string(reqEdn))

	req, err := http.NewRequest(http.MethodPost, ds.url, bytes.NewBuffer(reqEdn))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/edn")

	authToken := ds.token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if r.StatusCode >= 400 {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		body := buf.String()

		headers := ""
		if responseHeaderBytes, err := json.Marshal(req.Header); err != nil {
			headers = "Unable to read headers"
		} else {
			headers = string(responseHeaderBytes)
		}

		return nil, fmt.Errorf("async request returned unexpected status %s - HEADERS: %s BODY: %s", r.Status, headers, body)
	}

	return &QueryResponse{AsyncRequestMade: true}, nil
}
