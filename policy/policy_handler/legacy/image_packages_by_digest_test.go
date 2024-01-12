package legacy

import (
	"context"
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/internal/test_util"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockDs struct {
	t *testing.T
}

func (ds MockDs) Query(ctx context.Context, queryName string, query string, variables map[string]interface{}, output interface{}) (*data.QueryResponse, error) {
	assert.Equal(ds.t, queryName, vulnerabilitiesByPackageQueryName)
	assert.Equal(ds.t, query, vulnerabilitiesByPackageQuery)

	r := output.(*VulnerabilitiesByPackageResponse)
	r.VulnerabilitiesByPackage = []VulnerabilitiesByPackage{
		{
			Purl: "pkg:deb/ubuntu/libpcre3@2:8.39-12ubuntu0.1?arch=amd64&upstream=pcre3&distro=ubuntu-20.04",
			Vulnerabilities: []Vulnerability{{
				Cvss: Cvss{
					Severity: test_util.Pointer("HIGH"),
					Score:    test_util.Pointer(float32(7.5)),
				},
				FixedBy:         nil,
				PublishedAt:     "2017-07-10T11:29:00Z",
				Source:          "nist",
				SourceID:        "CVE-2017-11164",
				UpdatedAt:       "2023-04-12T11:15:00Z", // 2006-01-02T15:04:05Z07:00
				URL:             test_util.Pointer("https://scout.docker.com/v/CVE-2017-11164"),
				VulnerableRange: ">=0",
			}},
		},
	}

	return &data.QueryResponse{}, nil
}

func Test_mockImagePackagesByDigest(t *testing.T) {
	lPkgs := []Package{
		{
			Licenses:  []string{"GPL-3.0"},
			Name:      "libpcre3",
			Namespace: "pkgNamespace",
			Version:   "2:8.39-12ubuntu0.1",
			Purl:      "pkg:deb/ubuntu/libpcre3@2:8.39-12ubuntu0.1?arch=amd64&upstream=pcre3&distro=ubuntu-20.04",
			Type:      "pkgType",
		},
		{
			Licenses:  []string{"AGPL"},
			Name:      "coreutils",
			Namespace: "coreutilsNamespace",
			Version:   "8.30-3ubuntu2",
			Purl:      "pkg:deb/ubuntu/coreutils@8.30-3ubuntu2?arch=amd64&distro=ubuntu-20.04",
			Type:      "coreutilsType",
		},
	}

	actual, err := mockImagePackagesByDigest(context.TODO(), skill.RequestContext{}, lPkgs, MockDs{t})
	assert.NoError(t, err)

	expected := ImagePackagesByDigestResponse{
		ImagePackagesByDigest: &ImagePackagesByDigest{
			ImagePackages: ImagePackages{
				Packages: []Packages{
					{
						Package: PackageWithLicenses{
							Licenses:  []string{"GPL-3.0"},
							Name:      "libpcre3",
							Namespace: test_util.Pointer("pkgNamespace"),
							Version:   "2:8.39-12ubuntu0.1",
							Purl:      "pkg:deb/ubuntu/libpcre3@2:8.39-12ubuntu0.1?arch=amd64&upstream=pcre3&distro=ubuntu-20.04",
							Type:      "pkgType",
							Vulnerabilities: []Vulnerability{{
								Cvss: Cvss{
									Severity: test_util.Pointer("HIGH"),
									Score:    test_util.Pointer(float32(7.5)),
								},
								FixedBy:         nil,
								PublishedAt:     "2017-07-10T11:29:00Z",
								Source:          "nist",
								SourceID:        "CVE-2017-11164",
								UpdatedAt:       "2023-04-12T11:15:00Z", // 2006-01-02T15:04:05Z07:00
								URL:             test_util.Pointer("https://scout.docker.com/v/CVE-2017-11164"),
								VulnerableRange: ">=0",
							}},
						},
					},
					{
						Package: PackageWithLicenses{
							Licenses:        []string{"AGPL"},
							Name:            "coreutils",
							Namespace:       test_util.Pointer("coreutilsNamespace"),
							Version:         "8.30-3ubuntu2",
							Purl:            "pkg:deb/ubuntu/coreutils@8.30-3ubuntu2?arch=amd64&distro=ubuntu-20.04",
							Type:            "coreutilsType",
							Vulnerabilities: nil,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}
