package data

import (
	"github.com/atomist-skills/go-skill/policy/data/query"
)

type DataSource struct {
	jynxGQLClient query.QueryClient
}

func NewDataSource(graphQLClient query.QueryClient) DataSource {
	return DataSource{
		jynxGQLClient: graphQLClient,
	}
}

func (ds *DataSource) GetQueryClient() query.QueryClient {
	return ds.jynxGQLClient
}
