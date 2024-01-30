package legacy

import (
	"github.com/atomist-skills/go-skill/policy/types"
	"olympos.io/encoding/edn"
)

func BuildLocalEvalMocks(sb *types.SBOM) map[edn.Keyword]edn.RawMessage {
	m := map[edn.Keyword]edn.RawMessage{}
	if sb == nil {
		return m
	}

	m[ImagePackagesByDigestQueryName], _ = edn.Marshal(MockImagePackagesByDigestForLocalEval(sb))

	if sb.Source.Image != nil && sb.Source.Image.Config != nil {
		m[GetUserQueryName], _ = edn.Marshal(MockGetUserForLocalEval(sb.Source.Image.Config.Config.User))
	}

	if sb.Source.Provenance != nil {
		m[GetInTotoAttestationsQueryName], _ = edn.Marshal(MockGetInTotoAttestationsForLocalEval(sb))
	}

	return m
}

func MockImagePackagesByDigestForLocalEval(sb *types.SBOM) ImagePackagesByDigestResponse {
	vulns := map[string][]Vulnerability{}
	for _, tuple := range sb.Vulnerabilities {
		vulnsForPurl := []Vulnerability{}
		for _, v := range tuple.Vulnerabilities {
			vulnsForPurl = append(vulnsForPurl, Vulnerability{
				Cvss: Cvss{
					Severity: &v.Cvss.Severity,
					Score:    &v.Cvss.Score,
				},
				PublishedAt:     v.PublishedAt,
				UpdatedAt:       v.UpdatedAt,
				FixedBy:         &v.FixedBy,
				Source:          v.Source,
				SourceID:        v.SourceId,
				URL:             &v.Url,
				VulnerableRange: v.VulnerableRange,
			})
		}
		vulns[tuple.Purl] = vulnsForPurl
	}

	pkgs := []Packages{}
	for _, a := range sb.Artifacts {
		pkgs = append(pkgs, Packages{
			Package: PackageWithLicenses{
				Licenses:        a.Licenses,
				Name:            a.Name,
				Namespace:       &a.Namespace,
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
	}
}
