package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
	"golang.org/x/oauth2"
)

type ProxyClient struct {
	httpClient      http.Client
	correlationId   string
	gqlUrl          string
	entitlementsUrl string
}

func NewProxyClientFromSkillRequest(ctx context.Context, req skill.RequestContext) ProxyClient {
	return NewProxyClient(ctx, req.Event.Urls.Graphql, req.Event.Urls.Entitlements, req.Event.Token, req.Event.ExecutionId)
}

func NewProxyClient(ctx context.Context, graphqlUrl, entitlementsUrl, token, correlationId string) ProxyClient {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token, TokenType: "Bearer"},
	))

	return ProxyClient{
		httpClient:      *httpClient,
		correlationId:   correlationId,
		gqlUrl:          graphqlUrl,
		entitlementsUrl: entitlementsUrl,
	}
}

func (c *ProxyClient) Evaluate(ctx context.Context, organization, teamId, url string, sbom *types.SBOM, args map[string]interface{}) (goals.EvaluationResult, error) {
	preq := EvaluateRequest{
		EvaluateOptions: EvaluateOptions{
			Organization: organization,
			WorkspaceId:  teamId,
			Parameters:   args,
			URLs: struct {
				GraphQL      string `json:"graphql"`
				Entitlements string `json:"entitlements"`
			}{GraphQL: c.gqlUrl, Entitlements: c.entitlementsUrl},
		},
		SBOM: sbom,
	}

	data, err := json.Marshal(preq)
	if err != nil {
		return goals.EvaluationResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return goals.EvaluationResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Atomist-Correlation-Id", c.correlationId)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return goals.EvaluationResult{}, err
	}

	if res.StatusCode != http.StatusAccepted {
		return goals.EvaluationResult{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	defer res.Body.Close() //nolint:errcheck
	var resp EvaluateResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return goals.EvaluationResult{}, err
	}
	return resp.Result, nil
}
