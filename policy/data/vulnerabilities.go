package data

import (
	"context"

	"github.com/openvex/go-vex/pkg/vex"
	govex "github.com/openvex/go-vex/pkg/vex"

	"github.com/atomist-skills/go-skill/policy/data/query"
	"github.com/atomist-skills/go-skill/policy/data/query/jynx"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/atomist-skills/go-skill/sbom/normalization"
)

func (ds *DataSource) GetImageVulnerabilities(ctx context.Context, evalCtx goals.GoalEvaluationContext, imageSbom types.SBOM) (*query.QueryResponse, []types.Package, map[string][]types.Vulnerability, error) {
	var packages []types.Package
	vulns := map[string][]types.Vulnerability{}
	if len(imageSbom.Vulnerabilities) > 0 {
		for _, vulnsByPurl := range imageSbom.Vulnerabilities {
			vulns[vulnsByPurl.Purl] = vulnsByPurl.Vulnerabilities
		}

		packages = imageSbom.Artifacts
	} else if len(imageSbom.Artifacts) > 0 {
		packages = imageSbom.Artifacts

		evalCtx.Log.Debug("Normalizing purls from SBOM before fetching vulnerabilities")
		purls, purlMapping := normalization.NormalizeSBOM(&imageSbom)
		evalCtx.Log.Debugf("Normalized purls: %+v", purls)
		evalCtx.Log.Debugf("Purl mapping: %+v", purlMapping)

		var vulnsResponse types.VulnerabilitiesByPurls
		r, err := ds.jynxGQLClient.Query(ctx, jynx.VulnerabilitiesByPackageQueryName, jynx.VulnerabilitiesByPackageQuery, map[string]interface{}{
			"context": query.GqlContext(evalCtx),
			"purls":   purls,
			"digest":  imageSbom.Source.Image.Digest,
		}, &vulnsResponse)
		if err != nil || r.AsyncRequestMade {
			return r, nil, nil, err
		}

		evalCtx.Log.Debug("Denormalizing purls after fetching vulnerabilities")
		normalization.DenormalizeSBOM(&vulnsResponse, purlMapping)

		for _, vulnsByPurl := range vulnsResponse.VulnerabilitiesByPackage {
			vulns[vulnsByPurl.Purl] = applyVEX(vulnsByPurl, imageSbom.VexDocuments)
		}
	} else {
		var response jynx.ImagePackagesByDigestResponse
		r, err := ds.jynxGQLClient.Query(ctx, jynx.ImagePackagesByDigestQueryName, jynx.ImagePackagesByDigestQuery, map[string]interface{}{
			"context": query.GqlContext(evalCtx),
			"digest":  imageSbom.Source.Image.Digest,
		}, &response)
		if err != nil || r.AsyncRequestMade {
			return r, nil, nil, err
		}

		if response.ImagePackagesByDigest == nil {
			return r, nil, nil, nil
		}

		packages, vulns = convertGraphqlToPackages(*response.ImagePackagesByDigest)
	}

	return &query.QueryResponse{}, packages, vulns, nil
}

// applyVEX returns the CVEs that remain relevant after cross-referencing them with VEX documents.
func applyVEX(vulnsByPurl types.VulnerabilitiesByPurl, vexDocs []vex.VEX) []types.Vulnerability {
	filteredOutCVEs := []types.Vulnerability{}

	for _, cve := range vulnsByPurl.Vulnerabilities {
		for _, v := range vexDocs {
			for _, stmt := range v.Statements {
				if cveMatch(cve.SourceId, stmt) {
					if purlMatch(vulnsByPurl.Purl, stmt) {
						if notAffectedOrFixed(stmt) {
							filteredOutCVEs = append(filteredOutCVEs, cve)
						}
					}
				}
			}
		}
	}

	vexedCVEsMap := make(map[string]bool, len(filteredOutCVEs))
	for _, cve := range filteredOutCVEs {
		vexedCVEsMap[cve.SourceId] = true
	}

	// Filter out the VEXed CVEs
	cves := make([]types.Vulnerability, 0, len(vulnsByPurl.Vulnerabilities))
	for _, cve := range vulnsByPurl.Vulnerabilities {
		if !vexedCVEsMap[cve.SourceId] {
			cves = append(cves, cve)
		}
	}
	return cves
}

// cveMatch checks whether a CVE is present in a VEX statement
func cveMatch(cveID string, stmt govex.Statement) bool {
	return stmt.Vulnerability.ID == cveID || string(stmt.Vulnerability.Name) == cveID
}

// purlMatch checks whether a purl is present in at least one of the following locations:
// - Component
// - Subcomponent(s)
// - Special case for org-scoped VEXed CVEs.
func purlMatch(purl string, stmt govex.Statement) bool {
	purl, upstreamPurl := normalization.NormalizePURL(purl, nil)

	for _, p := range stmt.Products {
		// Check if purl is defined as the top-level component
		if purl == p.Component.ID {
			return true
		}
		// Check if purl is defined as one of the subcomponents
		if normalization.ContainsPurl(p.Subcomponents, purl) || normalization.ContainsPurl(p.Subcomponents, upstreamPurl) {
			return true
		}
		// If none of the previous conditions matched, we add this special case to support image-scoped exceptions.
		// The purpose of this is to align with how VEX works in the platform side.
		if len(p.Subcomponents) == 0 {
			return true
		}
	}

	return false
}

// notAffectedOrFixed checks whether the statement status is not affected or fixed.
func notAffectedOrFixed(stmt govex.Statement) bool {
	return stmt.Status == govex.StatusNotAffected || stmt.Status == govex.StatusFixed
}
