package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

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
		Data   map[edn.Keyword]edn.RawMessage `edn:"data"`
		Errors []struct {
			Message string `edn:"message"`
		}
	}

	AsyncDataSource struct {
		log      skill.Logger
		url      string
		token    string
		metadata string
	}
)

func NewAsyncDataSource(req skill.RequestContext, metadata string) AsyncDataSource {
	return AsyncDataSource{
		log:      req.Log,
		url:      fmt.Sprintf("%s:enqueue", req.Event.Urls.Graphql),
		token:    req.Event.Token,
		metadata: metadata,
	}
}

func (ds AsyncDataSource) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}) (*QueryResponse, error) {
	ednVariables := map[edn.Keyword]interface{}{}
	for k, v := range variables {
		ednVariables[edn.Keyword(k)] = v
	}

	request := AsyncQueryRequest{
		Name: queryName,
		Body: AsyncQueryBody{
			Query:     query,
			Variables: ednVariables,
		},
		Metadata: ds.metadata,
	}

	edn, err := edn.Marshal(request)
	if err != nil {
		return nil, err
	}

	ds.log.Infof("Async request: %s", string(edn))

	req, err := http.NewRequest(http.MethodPost, ds.url, bytes.NewBuffer(edn))
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

func UnwrapAsyncResponse(result map[edn.Keyword]edn.RawMessage) (DataSource, error) {
	ednBody, err := edn.Marshal(result)
	if err != nil {
		return nil, err
	}

	var response AsyncQueryResponse
	err = edn.Unmarshal(ednBody, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf(response.Errors[0].Message)
	}

	queryResponses := map[string][]byte{}
	for k, v := range response.Data {
		queryResponses[string(k)] = v
	}

	return NewFixedDataSource(queryResponses), nil
}
