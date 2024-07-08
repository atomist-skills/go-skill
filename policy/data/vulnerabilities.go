package data

import (
	"context"

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
			affected := true
			for _, v := range imageSbom.VexDocuments {
				for _, stmt := range v.Statements {
					purl, upstreamPurl := normalization.NormalizePURL(vulnsByPurl.Purl, nil)
					for _, p := range stmt.Products {
						if normalization.ContainsPurl(p.Subcomponents, purl) || normalization.ContainsPurl(p.Subcomponents, upstreamPurl) {
							if stmt.Status == govex.StatusNotAffected || stmt.Status == govex.StatusFixed {
								affected = false
							}
						}
					}
				}
			}

			if affected {
				vulns[vulnsByPurl.Purl] = vulnsByPurl.Vulnerabilities
			}
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
