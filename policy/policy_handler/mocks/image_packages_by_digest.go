package mocks

import (
	"context"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/policy_handler/legacy"
	"github.com/atomist-skills/go-skill/policy/types"
)

func MockImagePackagesByDigest(ctx context.Context, req skill.RequestContext, sb *types.SBOM) (legacy.ImagePackagesByDigestResponse, error) {
	req.Log.Info("Building local evaluation mocks for image packages by digest")

	if len(sb.Vulnerabilities) == 0 {
		req.Log.Info("SBOM doesn't provide any vulnerabilities directly, fetching them from GraphQL")

		var pkgs []legacy.Package
		for _, a := range sb.Artifacts {
			pkgs = append(pkgs, legacy.Package{
				Type:      a.Type,
				Namespace: a.Namespace,
				Name:      a.Name,
				Version:   a.Version,
				Purl:      a.Purl,
				Licenses:  a.Licenses,
			})
		}
		res, err := legacy.MockImagePackagesByDigest(ctx, req, pkgs, sb)
		if err != nil {
			return res, err
		}
		return res, nil
	}

	req.Log.Info("SBOM provides vulnerabilities")
	vulns := map[string][]legacy.Vulnerability{}
	for _, tuple := range sb.Vulnerabilities {
		vulnsForPurl := []legacy.Vulnerability{}
		for _, v := range tuple.Vulnerabilities {
			vulnsForPurl = append(vulnsForPurl, legacy.Vulnerability{
				Cvss: legacy.Cvss{
					Severity: &v.Cvss.Severity,
					Score:    &v.Cvss.Score,
				},
				PublishedAt:     v.PublishedAt,
				UpdatedAt:       v.UpdatedAt,
				FixedBy:         &v.FixedBy,
				Source:          v.Source,
				SourceId:        v.SourceId,
				URL:             &v.Url,
				VulnerableRange: v.VulnerableRange,
			})
		}
		vulns[tuple.Purl] = vulnsForPurl
	}

	pkgs := []legacy.Packages{}
	for _, a := range sb.Artifacts {
		pkgs = append(pkgs, legacy.Packages{
			Package: legacy.PackageWithLicenses{
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

	return legacy.ImagePackagesByDigestResponse{
		ImagePackagesByDigest: &legacy.ImagePackagesByDigest{
			ImagePackages: legacy.ImagePackages{
				Packages: pkgs,
			},
		},
	}, nil
}
