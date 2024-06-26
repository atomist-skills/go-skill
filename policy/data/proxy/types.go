package proxy

import (
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
)

type EvaluateRequest struct {
	EvaluateOptions
	SBOM *types.SBOM `json:"sbom"`
}

type EvaluateResponse struct {
	Result goals.EvaluationResult `json:"result"`
}

type EvaluateOptions struct {
	Organization string                 `json:"organization"`
	WorkspaceId  string                 `json:"workspaceId"`
	Parameters   map[string]interface{} `json:"parameters"`
	URLs         struct {
		GraphQL      string `json:"graphql"`
		Entitlements string `json:"entitlements"`
	}
}
