package data

import (
	"context"
	"testing"

	"github.com/atomist-skills/go-skill/internal/test_util"
	"github.com/atomist-skills/go-skill/policy/data/query"
	"github.com/atomist-skills/go-skill/policy/data/query/jynx"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/stretchr/testify/assert"
)

type VulnTestQueryClient struct {
	vulnsByPurls     types.VulnerabilitiesByPurls
	packagesResponse jynx.ImagePackagesByDigestResponse
}

func NewVulnTestQueryClient(vulnsByPurls types.VulnerabilitiesByPurls, packagesResponse jynx.ImagePackagesByDigestResponse) VulnTestQueryClient {
	return VulnTestQueryClient{
		vulnsByPurls:     vulnsByPurls,
		packagesResponse: packagesResponse,
	}
}

func (ds VulnTestQueryClient) Query(ctx context.Context, queryName string, queryBody string, variables map[string]interface{}, output interface{}) (*query.QueryResponse, error) {
	if queryName == jynx.VulnerabilitiesByPackageQueryName {
		output.(*types.VulnerabilitiesByPurls).VulnerabilitiesByPackage = ds.vulnsByPurls.VulnerabilitiesByPackage
	} else if queryName == jynx.ImagePackagesByDigestQueryName {
		output.(*jynx.ImagePackagesByDigestResponse).ImagePackagesByDigest = ds.packagesResponse.ImagePackagesByDigest
	}

	return &query.QueryResponse{}, nil
}

func Test_GetImageVulnerabilities_WhenSbomHasVulnerabilities(t *testing.T) {
	sbom := types.SBOM{
		Vulnerabilities: []types.VulnerabilitiesByPurl{
			{
				Purl: "pkg:pypi/requests@2.25.1",
				Vulnerabilities: []types.Vulnerability{
					{
						SourceId: "CVE-2021-3456",
						Cvss: types.Cvss{
							Score:    9.8,
							Severity: "CRITICAL",
						},
					},
					{
						SourceId: "CVE-2022-1226",
						Cvss: types.Cvss{
							Score:    7.5,
							Severity: "HIGH",
						},
					},
				},
			},
			{
				Purl: "pkg:npm/my-package@1.2.3",
				Vulnerabilities: []types.Vulnerability{
					{
						SourceId: "CVE-2021-2256",
						FixedBy:  "1.2.4",
						Cvss: types.Cvss{
							Score:    5.6,
							Severity: "MEDIUM",
						},
					},
				},
			},
		},
		Artifacts: []types.Package{
			{
				Purl: "pkg:pypi/requests@2.25.1",
			},
			{
				Purl: "pkg:npm/my-package@1.2.3",
			},
		},
	}

	expectedPackages := []types.Package{
		{
			Purl: "pkg:pypi/requests@2.25.1",
		},
		{
			Purl: "pkg:npm/my-package@1.2.3",
		},
	}

	expectedVulnerabilities := map[string][]types.Vulnerability{
		"pkg:pypi/requests@2.25.1": {
			{
				SourceId: "CVE-2021-3456",
				Cvss: types.Cvss{
					Score:    9.8,
					Severity: "CRITICAL",
				},
			},
			{
				SourceId: "CVE-2022-1226",
				Cvss: types.Cvss{
					Score:    7.5,
					Severity: "HIGH",
				},
			},
		},
		"pkg:npm/my-package@1.2.3": {
			{
				SourceId: "CVE-2021-2256",
				FixedBy:  "1.2.4",
				Cvss: types.Cvss{
					Score:    5.6,
					Severity: "MEDIUM",
				},
			},
		},
	}

	ds := DataSource{}

	response, packages, vulnerabilities, err := ds.GetImageVulnerabilities(context.Background(), goals.GoalEvaluationContext{}, sbom)

	assert.Nil(t, err)
	assert.False(t, response.AsyncRequestMade)

	assert.Equal(t, expectedPackages, packages)
	assert.Equal(t, expectedVulnerabilities, vulnerabilities)
}

func Test_GetImageVulnerabilities_WhenSbomHasArtifacts_AndNoVulnerabilities(t *testing.T) {
	sbom := types.SBOM{
		Source: types.Source{
			Image: &types.ImageSource{
				Digest: "sha256:123456",
			},
		},
		Artifacts: []types.Package{
			{
				Purl: "pkg:pypi/requests@2.25.1",
			},
			{
				Purl: "pkg:npm/my-package@1.2.3",
			},
		},
	}

	expectedPackages := []types.Package{
		{
			Purl: "pkg:pypi/requests@2.25.1",
		},
		{
			Purl: "pkg:npm/my-package@1.2.3",
		},
	}

	expectedVulnerabilities := map[string][]types.Vulnerability{
		"pkg:pypi/requests@2.25.1": {
			{
				SourceId: "CVE-2021-3456",
				Cvss: types.Cvss{
					Score:    9.8,
					Severity: "CRITICAL",
				},
			},
			{
				SourceId: "CVE-2022-1226",
				Cvss: types.Cvss{
					Score:    7.5,
					Severity: "HIGH",
				},
			},
		},
		"pkg:npm/my-package@1.2.3": {
			{
				SourceId: "CVE-2021-2256",
				FixedBy:  "1.2.4",
				Cvss: types.Cvss{
					Score:    5.6,
					Severity: "MEDIUM",
				},
			},
		},
	}

	ds := DataSource{
		jynxGQLClient: NewVulnTestQueryClient(types.VulnerabilitiesByPurls{
			VulnerabilitiesByPackage: []types.VulnerabilitiesByPurl{
				{
					Purl: "pkg:pypi/requests@2.25.1",
					Vulnerabilities: []types.Vulnerability{
						{
							SourceId: "CVE-2021-3456",
							Cvss: types.Cvss{
								Score:    9.8,
								Severity: "CRITICAL",
							},
						},
						{
							SourceId: "CVE-2022-1226",
							Cvss: types.Cvss{
								Score:    7.5,
								Severity: "HIGH",
							},
						},
					},
				},
				{
					Purl: "pkg:npm/my-package@1.2.3",
					Vulnerabilities: []types.Vulnerability{
						{
							SourceId: "CVE-2021-2256",
							FixedBy:  "1.2.4",
							Cvss: types.Cvss{
								Score:    5.6,
								Severity: "MEDIUM",
							},
						},
					},
				},
			},
		},
			jynx.ImagePackagesByDigestResponse{}),
	}

	response, packages, vulnerabilities, err := ds.GetImageVulnerabilities(context.Background(), goals.GoalEvaluationContext{Log: test_util.CreateEmptyLogger()}, sbom)

	assert.Nil(t, err)
	assert.False(t, response.AsyncRequestMade)

	assert.Equal(t, expectedPackages, packages)
	assert.Equal(t, expectedVulnerabilities, vulnerabilities)
}

func Test_GetImageVulnerabilities_WhenSbomHasNoArtifacts_AndNoVulnerabilities(t *testing.T) {
	sbom := types.SBOM{
		Source: types.Source{
			Image: &types.ImageSource{
				Digest: "sha256:123456",
			},
		},
	}

	expectedPackages := []types.Package{
		{
			Purl: "pkg:pypi/requests@2.25.1",
		},
		{
			Purl: "pkg:npm/my-package@1.2.3",
		},
	}

	expectedVulnerabilities := map[string][]types.Vulnerability{
		"pkg:pypi/requests@2.25.1": {
			{
				SourceId: "CVE-2021-3456",
				Cvss: types.Cvss{
					Score:    9.8,
					Severity: "CRITICAL",
				},
			},
			{
				SourceId: "CVE-2022-1226",
				Cvss: types.Cvss{
					Score:    7.5,
					Severity: "HIGH",
				},
			},
		},
		"pkg:npm/my-package@1.2.3": {
			{
				SourceId: "CVE-2021-2256",
				FixedBy:  "1.2.4",
				Cvss: types.Cvss{
					Score:    5.6,
					Severity: "MEDIUM",
				},
			},
		},
	}

	ds := DataSource{
		jynxGQLClient: NewVulnTestQueryClient(
			types.VulnerabilitiesByPurls{},
			jynx.ImagePackagesByDigestResponse{
				ImagePackagesByDigest: &jynx.ImagePackagesByDigest{
					ImagePackages: jynx.ImagePackages{
						Packages: []jynx.Packages{
							{
								Package: jynx.Package{
									Purl: "pkg:pypi/requests@2.25.1",
									Vulnerabilities: []jynx.Vulnerability{
										{
											SourceID: "CVE-2021-3456",
											Cvss: jynx.Cvss{
												Score:    Ptr(float32(9.8)),
												Severity: Ptr("CRITICAL"),
											},
										},
										{
											SourceID: "CVE-2022-1226",
											Cvss: jynx.Cvss{
												Score:    Ptr(float32(7.5)),
												Severity: Ptr("HIGH"),
											},
										},
									},
								},
							},
							{
								Package: jynx.Package{
									Purl: "pkg:npm/my-package@1.2.3",
									Vulnerabilities: []jynx.Vulnerability{
										{
											SourceID: "CVE-2021-2256",
											FixedBy:  Ptr("1.2.4"),
											Cvss: jynx.Cvss{
												Score:    Ptr(float32(5.6)),
												Severity: Ptr("MEDIUM"),
											},
										},
									},
								},
							},
						},
					},
				},
			}),
	}

	response, packages, vulnerabilities, err := ds.GetImageVulnerabilities(context.Background(), goals.GoalEvaluationContext{Log: test_util.CreateEmptyLogger()}, sbom)

	assert.Nil(t, err)
	assert.False(t, response.AsyncRequestMade)

	assert.Equal(t, expectedPackages, packages)
	assert.Equal(t, expectedVulnerabilities, vulnerabilities)
}

func Ptr[T any](v T) *T {
	return &v
}
