package legacy

import (
	"context"
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
)

// Versions of scout-cli-plugin created before the introduction of fixedQueryResults
// directly passed a []Package object in the metadata for local evaluation.
// This was then supplemented by a synchronous GraphQL call to load vulnerability data,
// so we mock the entire process to support these older versions.
// TODO remove this whole system when no longer used

const (
	ImagePackagesByDigestQueryName    = "image-packages-by-digest"
	vulnerabilitiesByPackageQueryName = "vulnerabilities-by-package"

	// language=graphql
	vulnerabilitiesByPackageQuery = `
	query ($context: Context!, $packageUrls: [String!]!) {
		vulnerabilitiesByPackage(context: $context, packageUrls: $packageUrls) {
			purl
			vulnerabilities {
			cvss {
				severity
				score
			}
			fixedBy
			publishedAt
			source
			sourceId
			updatedAt
			url
			vulnerableRange
			}
		}
	}`
)

type (
	Package struct {
		Licenses  []string `edn:"licenses,omitempty"` // only needed for the license policy evaluation
		Name      string   `edn:"name"`
		Namespace string   `edn:"namespace"`
		Version   string   `edn:"version"`
		Purl      string   `edn:"purl"`
		Type      string   `edn:"type"`
	}

	ImagePackagesByDigestResponse struct {
		ImagePackagesByDigest *ImagePackagesByDigest `json:"imagePackagesByDigest" edn:"imagePackagesByDigest"`
	}

	ImagePackagesByDigest struct {
		ImagePackages ImagePackages `json:"imagePackages" edn:"imagePackages"`
	}

	ImagePackages struct {
		Packages []Packages `json:"packages" edn:"packages"`
	}

	Packages struct {
		Package PackageWithLicenses `json:"package" edn:"package"`
	}

	PackageWithLicenses struct {
		Licenses        []string        `json:"licenses" edn:"licenses"`
		Name            string          `json:"name" edn:"name"`
		Namespace       *string         `json:"namespace" edn:"namespace"`
		Version         string          `json:"version" edn:"version"`
		Purl            string          `json:"purl" edn:"purl"`
		Type            string          `json:"type" edn:"type"`
		Vulnerabilities []Vulnerability `json:"vulnerabilities" edn:"vulnerabilities"`
	}

	Vulnerability struct {
		Cvss            Cvss    `json:"cvss"`
		FixedBy         *string `json:"fixedBy"`
		PublishedAt     string  `json:"publishedAt"`
		Source          string  `json:"source"`
		SourceID        string  `json:"sourceId"`
		UpdatedAt       string  `json:"updatedAt"`
		URL             *string `json:"url"`
		VulnerableRange string  `json:"vulnerableRange"`
	}

	Cvss struct {
		Severity *string  `json:"severity"`
		Score    *float32 `json:"score"`
	}

	VulnerabilitiesByPackageResponse struct {
		VulnerabilitiesByPackage []VulnerabilitiesByPackage `json:"vulnerabilitiesByPackage"`
	}

	VulnerabilitiesByPackage struct {
		Purl            string          `json:"purl"`
		Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	}
)

func MockImagePackagesByDigest(ctx context.Context, req skill.RequestContext, sbomPkgs []Package) (ImagePackagesByDigestResponse, error) {
	// separated for testing
	ds, err := data.NewSyncGraphqlDataSource(ctx, req)
	if err != nil {
		return ImagePackagesByDigestResponse{}, err
	}

	return mockImagePackagesByDigest(ctx, req, sbomPkgs, ds)
}

func mockImagePackagesByDigest(ctx context.Context, req skill.RequestContext, sbomPkgs []Package, ds data.DataSource) (ImagePackagesByDigestResponse, error) {
	purls := []string{}
	for _, p := range sbomPkgs {
		purls = append(purls, p.Purl)
	}

	var vulnsResponse VulnerabilitiesByPackageResponse
	_, err := ds.Query(ctx, vulnerabilitiesByPackageQueryName, vulnerabilitiesByPackageQuery, map[string]interface{}{
		"context":     data.GqlContext(req),
		"packageUrls": purls,
	}, &vulnsResponse)
	if err != nil {
		return ImagePackagesByDigestResponse{}, err
	}

	vulns := map[string][]Vulnerability{}
	for _, v := range vulnsResponse.VulnerabilitiesByPackage {
		vulns[v.Purl] = v.Vulnerabilities
	}

	pkgs := []Packages{}
	for _, a := range sbomPkgs {
		ns := a.Namespace
		pkgs = append(pkgs, Packages{
			Package: PackageWithLicenses{
				Licenses:        a.Licenses,
				Name:            a.Name,
				Namespace:       &ns,
				Version:         a.Version,
				Purl:            a.Purl,
				Type:            a.Type,
				Vulnerabilities: vulns[a.Purl],
			},
		})
	}

	return ImagePackagesByDigestResponse{
		ImagePackagesByDigest: &ImagePackagesByDigest{
			ImagePackages: ImagePackages{
				Packages: pkgs,
			},
		},
	}, nil
}
