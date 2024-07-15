package data

import (
	"context"
	"testing"

	govex "github.com/openvex/go-vex/pkg/vex"

	"github.com/atomist-skills/go-skill/internal/test_util"
	"github.com/atomist-skills/go-skill/policy/data/query"
	"github.com/atomist-skills/go-skill/policy/data/query/jynx"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/openvex/go-vex/pkg/vex"
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

func Test_applyVEX(t *testing.T) {
	const (
		openSSLPurl   = "pkg:apk/alpine/openssl@3.0.12-r1?os_name=alpine&os_version=3.17"
		alpineImgPurl = "pkg:docker/alpine@sha256:6e94b5cda2d6fd57d85abf81e81dabaea97a5885f919da676cc19d3551da4061"
		awsPurl       = "pkg:golang/github.com/aws/aws-sdk-go@1.44.288"
	)

	tests := []struct {
		name         string
		vulnsByPurl  types.VulnerabilitiesByPurl
		vexDocs      []vex.VEX
		expectedCVEs []types.Vulnerability // CVEs after applying VEX
	}{
		{
			name: "CVE-2024-5535 is not filtered out when there aren't VEX documents",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535"),
			},
			vexDocs:      []vex.VEX{}, // empty on purpose
			expectedCVEs: cves("CVE-2024-5535"),
		},
		{
			name: "CVE-2024-5535 is not filtered out when the VEX document has no statements",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{}, // empty on purpose
				},
			},
			expectedCVEs: cves("CVE-2024-5535"),
		},
		{
			name: "CVE-2024-5535 is not filtered out when purl is not present in either the product id or subcomponents",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								ID: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: alpineImgPurl,
									},
									Subcomponents: []vex.Subcomponent{
										{
											Component: vex.Component{
												ID: awsPurl,
											},
										},
									},
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: cves("CVE-2024-5535"),
		},
		{
			name: "CVE-2024-5535 is filtered out when purl matches the product id",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535", "CVE-2024-5536"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								ID: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: openSSLPurl,
									},
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: cves("CVE-2024-5536"),
		},
		{
			name: "CVE-2024-5535 is filtered out when purl is present in subcomponents",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535", "CVE-2024-5536"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								ID: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: alpineImgPurl,
									},
									Subcomponents: []vex.Subcomponent{
										{
											Component: vex.Component{
												ID: openSSLPurl,
											},
										},
									},
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: cves("CVE-2024-5536"),
		},
		{
			name: "CVE-2024-5535 is filtered out when there are no subcomponents (even if there is a product id mismatch)",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535", "CVE-2024-5536"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								ID: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: alpineImgPurl, // notice product id mismatch with openSSLPurl
									},
									Subcomponents: []vex.Subcomponent{}, // empty on purpose
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: cves("CVE-2024-5536"),
		},
		{
			name: "CVE-2024-0001 is not filtered out when its source id does not match the vulnerability id in the VEX statement",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-0001"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								ID: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: alpineImgPurl,
									},
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: cves("CVE-2024-0001"),
		},
		{
			name: "CVE-2024-0001 is not filtered out when its source id does not match the vulnerability name in the VEX statement",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-0001"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								Name: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: alpineImgPurl,
									},
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: cves("CVE-2024-0001"),
		},
		{
			name: "CVE-2024-5535 is filtered out when status is not_affected",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								Name: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: openSSLPurl,
									},
								},
							},
							Status:        govex.StatusNotAffected,
							Justification: vex.VulnerableCodeNotInExecutePath,
						},
					},
				},
			},
			expectedCVEs: []types.Vulnerability{},
		},
		{
			name: "CVE-2024-5535 is filtered out when status is fixed",
			vulnsByPurl: types.VulnerabilitiesByPurl{
				Purl:            openSSLPurl,
				Vulnerabilities: cves("CVE-2024-5535"),
			},
			vexDocs: []vex.VEX{
				{
					Statements: []vex.Statement{
						{
							Vulnerability: vex.Vulnerability{
								Name: "CVE-2024-5535",
							},
							Products: []vex.Product{
								{
									Component: govex.Component{
										ID: openSSLPurl,
									},
								},
							},
							Status: govex.StatusFixed,
						},
					},
				},
			},
			expectedCVEs: []types.Vulnerability{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := applyVEX(tt.vulnsByPurl, tt.vexDocs)
			if len(actual) != len(tt.expectedCVEs) {
				t.Errorf("applyVEX() = %d, want %d", len(actual), len(tt.expectedCVEs))
			}
			if len(actual) == len(tt.expectedCVEs) {
				for i, v := range actual {
					if tt.expectedCVEs[i].SourceId != v.SourceId {
						t.Errorf("applyVEX() = %v, want %v", v.SourceId, tt.expectedCVEs[i].SourceId)
					}
				}
			}
		})
	}
}

func cves(cveIDs ...string) []types.Vulnerability {
	var cves = make([]types.Vulnerability, 0, len(cveIDs))
	for _, cve := range cveIDs {
		cves = append(cves, types.Vulnerability{
			SourceId: cve,
		})
	}
	return cves
}

func Ptr[T any](v T) *T {
	return &v
}
