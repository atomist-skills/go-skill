package data

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atomist-skills/go-skill"
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
		SubscriptionResults [][]edn.RawMessage            `edn:"subscription"`
		AsyncQueryResults   map[string]AsyncQueryResponse `edn:"results"`
		InFlightQueryName   string                        `edn:"query-name"`
	}

	AsyncDataSource struct {
		log                 skill.Logger
		url                 string
		token               string
		subscriptionResults [][]edn.RawMessage
		asyncResults        map[string]AsyncQueryResponse
	}
)

func NewAsyncDataSource(req skill.RequestContext, subscriptionResults [][]edn.RawMessage, asyncResults map[string]AsyncQueryResponse) AsyncDataSource {
	return AsyncDataSource{
		log:                 req.Log,
		url:                 fmt.Sprintf("%s:enqueue", req.Event.Urls.Graphql),
		token:               req.Event.Token,
		subscriptionResults: subscriptionResults,
		asyncResults:        asyncResults,
	}
}

func (ds AsyncDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*QueryResponse, error) {
	if existingResult, ok := ds.asyncResults[queryName]; ok {
		if len(existingResult.Errors) != 0 {
			return nil, fmt.Errorf("async query returned error: %s", existingResult.Errors[0].Message)
		}
		return &QueryResponse{}, edn.Unmarshal(existingResult.Data, output)
	}

	ednVariables := map[edn.Keyword]interface{}{}
	for k, v := range variables {
		ednVariables[edn.Keyword(k)] = v
	}

	metadata := AsyncResultMetadata{
		SubscriptionResults: ds.subscriptionResults,
		AsyncQueryResults:   ds.asyncResults,
		InFlightQueryName:   queryName,
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

		return nil, fmt.Errorf("async request returned unexpected status %s: %s", r.Status, body)
	}

	return &QueryResponse{AsyncRequestMade: true}, nil
}
