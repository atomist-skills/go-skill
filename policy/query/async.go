package query

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

type AsyncQueryRequest struct {
	Name     string      `edn:"name"`
	Body     interface{} `edn:"body"`
	Metadata string      `edn:"metadata"`
}

type AsyncQueryClient struct {
	log      skill.Logger
	token    string
	metadata string
}

func NewAsyncQueryClient(log skill.Logger, token string, metadata string) AsyncQueryClient {
	return AsyncQueryClient{
		log:      log,
		token:    token,
		metadata: metadata,
	}
}

func (c *AsyncQueryClient) SubmitAsyncQuery(name string, url string, body interface{}) error {
	request := AsyncQueryRequest{
		Name:     name,
		Body:     body,
		Metadata: c.metadata,
	}

	edn, err := edn.Marshal(request)
	if err != nil {
		return err
	}

	c.log.Infof("Async request: %s", string(edn))

	url = url + ":enqueue"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(edn))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/edn")

	authToken := c.token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if r.StatusCode >= 400 {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		body := buf.String()

		return fmt.Errorf("async request returned unexpected status %s: %s", r.Status, body)
	}

	return nil
}
