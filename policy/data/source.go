package data

import "github.com/atomist-skills/go-skill/policy/data/query"

type DataSource struct {
	graphQLClient query.QueryClient
}

func NewDataSource(graphQLClient query.QueryClient) DataSource {
	return DataSource{
		graphQLClient: graphQLClient,
	}
}

func (ds *DataSource) GetQueryClient() query.QueryClient {
	return ds.graphQLClient
}
