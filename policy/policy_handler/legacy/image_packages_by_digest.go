package legacy

import (
	"context"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/atomist-skills/go-skill/sbom/normalization"
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
	query ($context: Context!, $purls: [String!]!) {
		vulnerabilitiesByPackage(context: $context, packageUrls: $purls) {
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
		SourceId        string  `json:"sourceId"`
		UpdatedAt       string  `json:"updatedAt"`
		URL             *string `json:"url"`
		VulnerableRange string  `json:"vulnerableRange"`
	}

	Cvss struct {
		Severity *string  `json:"severity"`
		Score    *float32 `json:"score"`
	}
)

func MockImagePackagesByDigest(ctx context.Context, req skill.RequestContext, legacyPkgs []Package, sb *types.SBOM) (ImagePackagesByDigestResponse, error) {
	// separated for testing
	ds, err := data.NewSyncGraphqlDataSource(ctx, req)
	if err != nil {
		return ImagePackagesByDigestResponse{}, err
	}

	return mockImagePackagesByDigest(ctx, req, legacyPkgs, ds, sb)
}

func mockImagePackagesByDigest(ctx context.Context, req skill.RequestContext, legacyPkgs []Package, ds data.DataSource, sb *types.SBOM) (ImagePackagesByDigestResponse, error) {
	purls := []string{}
	purlMapping := map[string]string{}

	if sb != nil {
		req.Log.Debug("Normalizing purls from SBOM before fetching vulnerabilities")
		purls, purlMapping = normalization.NormalizeSBOM(sb)
		req.Log.Debugf("Normalized purls: %+v", purls)
		req.Log.Debugf("Purl mapping: %+v", purlMapping)
	} else {
		req.Log.Debug("Using packages from metadata (legacy) for fetching vulnerabilities")
		for _, p := range legacyPkgs {
			purls = append(purls, p.Purl)
		}
	}

	var vulnsResponse types.VulnerabilitiesByPurls
	_, err := ds.Query(ctx, vulnerabilitiesByPackageQueryName, vulnerabilitiesByPackageQuery, map[string]interface{}{
		"context": data.GqlContext(req),
		"purls":   purls,
	}, &vulnsResponse)
	if err != nil {
		return ImagePackagesByDigestResponse{}, err
	}

	if sb != nil {
		req.Log.Debug("Denormalizing purls after fetching vulnerabilities")
		normalization.DenormalizeSBOM(&vulnsResponse, purlMapping)
	}

	m := map[string][]Vulnerability{}
	for _, v := range vulnsResponse.VulnerabilitiesByPackage {
		vulns := []Vulnerability{}
		for _, vv := range v.Vulnerabilities {
			vuln := Vulnerability{
				Cvss: Cvss{
					Severity: &vv.Cvss.Severity,
					Score:    &vv.Cvss.Score,
				},
				PublishedAt:     vv.PublishedAt,
				Source:          vv.Source,
				SourceId:        vv.SourceId,
				UpdatedAt:       vv.UpdatedAt,
				URL:             &vv.Url,
				VulnerableRange: vv.VulnerableRange,
			}
			if vv.FixedBy != "" {
				vuln.FixedBy = &vv.FixedBy
			}
			vulns = append(vulns, vuln)
		}
		m[v.Purl] = vulns
	}

	pkgs := []Packages{}
	if sb != nil {
		// Build pkgs from SBOM artifacts
		req.Log.Debug("Building packages from SBOM artifacts")
		for _, a := range sb.Artifacts {
			ns := a.Namespace
			pkgs = append(pkgs, Packages{
				Package: PackageWithLicenses{
					Licenses:        a.Licenses,
					Name:            a.Name,
					Namespace:       &ns,
					Version:         a.Version,
					Purl:            a.Purl,
					Type:            a.Type,
					Vulnerabilities: m[a.Purl],
				},
			})
		}
	} else {
		// Build pkgs from input packages (legacy)
		req.Log.Debug("Building packages from metadata (legacy)")
		for _, a := range legacyPkgs {
			ns := a.Namespace
			pkgs = append(pkgs, Packages{
				Package: PackageWithLicenses{
					Licenses:        a.Licenses,
					Name:            a.Name,
					Namespace:       &ns,
					Version:         a.Version,
					Purl:            a.Purl,
					Type:            a.Type,
					Vulnerabilities: m[a.Purl],
				},
			})
		}
	}
	req.Log.Debugf("Returning %d packages", len(pkgs))

	res := ImagePackagesByDigestResponse{
		ImagePackagesByDigest: &ImagePackagesByDigest{
			ImagePackages: ImagePackages{
				Packages: pkgs,
			},
		},
	}
	req.Log.Debugf("Mocked ImagePackagesByDigestResponse: %+v", res)
	return res, nil
}
