package graphql

import (
	"context"
	"encoding/json"
)

type VulnerabilitiesByPackageResponse struct {
	VulnerabilitiesByPackage []VulnerabilitiesByPackage `json:"vulnerabilitiesByPackage"`
}

type VulnerabilitiesByPackage struct {
	Purl            string          `json:"purl"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

func (client *GraphqlSkillClient) GetVulnerabilitiesByPackage(ctx context.Context, purls []string) ([]VulnerabilitiesByPackage, error) {
	log := client.RequestContext.Log

	variables := map[string]interface{}{
		"context":     gqlContext(client),
		"packageUrls": purls,
	}

	log.Infof("Graphql endpoint: %s", client.RequestContext.Event.Urls.Graphql)
	log.Infof("Executing query: %s", vulnerabilitiesByPackageQuery)
	log.Infof("Query variables: %v", variables)

	res, err := client.GraphqlClient.ExecRaw(ctx, vulnerabilitiesByPackageQuery, variables)
	if err != nil {
		return nil, err
	}

	log.Infof("GraphQL query response: %s", string(res))

	var r VulnerabilitiesByPackageResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, err
	}

	return r.VulnerabilitiesByPackage, nil
}
