package data

import (
	"github.com/atomist-skills/go-skill/policy/data/query/jynx"
	"github.com/atomist-skills/go-skill/policy/types"
)

func convertGraphqlToPackages(imagePackages jynx.ImagePackagesByDigest) ([]types.Package, map[string][]types.Vulnerability) {
	var nonEmptyHistories []jynx.ImageHistory
	for _, history := range imagePackages.ImageHistories {
		if !history.EmptyLayer {
			nonEmptyHistories = append(nonEmptyHistories, history)
		}
	}

	var pkgs []types.Package
	var vulns = map[string][]types.Vulnerability{}
	for _, p := range imagePackages.ImagePackages.Packages {
		var locations []types.Location
		for _, location := range p.Locations {
			layerOrdinal := -1
			for _, layer := range imagePackages.ImageLayers.Layers {
				if location.DiffId == layer.DiffId {
					layerOrdinal = layer.Ordinal
					break
				}
			}

			historyOrdinal := -1
			if len(nonEmptyHistories) > 0 && layerOrdinal > -1 {
				historyOrdinal = nonEmptyHistories[layerOrdinal].Ordinal
			}

			locations = append(locations, types.Location{
				Ordinal: historyOrdinal,
				Path:    location.Path,
			})
		}

		var namespace string
		if p.Package.Namespace == nil {
			namespace = ""
		} else {
			namespace = *p.Package.Namespace
		}

		vulnerabilities := convertToVulnerabilities(p.Package.Vulnerabilities)

		pkgs = append(pkgs, types.Package{
			Purl:      p.Package.Purl,
			Licenses:  p.Package.Licenses,
			Name:      p.Package.Name,
			Namespace: namespace,
			Version:   p.Package.Version,
			Locations: locations,
		})

		vulns[p.Package.Purl] = vulnerabilities
	}

	return pkgs, vulns
}

func convertToVulnerabilities(vulnerabilities []jynx.Vulnerability) []types.Vulnerability {
	var result []types.Vulnerability

	for _, vulnerability := range vulnerabilities {
		vulnerabilityResult := types.Vulnerability{
			Cvss:            types.Cvss{},
			PublishedAt:     vulnerability.PublishedAt,
			Source:          vulnerability.Source,
			SourceId:        vulnerability.SourceID,
			UpdatedAt:       vulnerability.UpdatedAt,
			VulnerableRange: vulnerability.VulnerableRange,
			CisaExploited:   vulnerability.CisaExploited,
		}

		if vulnerability.Cvss.Score != nil {
			vulnerabilityResult.Cvss.Score = *vulnerability.Cvss.Score
		}

		if vulnerability.Cvss.Severity != nil {
			vulnerabilityResult.Cvss.Severity = *vulnerability.Cvss.Severity
		}

		if vulnerability.URL != nil {
			vulnerabilityResult.Url = *vulnerability.URL
		}

		if vulnerability.FixedBy != nil {
			vulnerabilityResult.FixedBy = *vulnerability.FixedBy
		}

		if vulnerability.Epss != nil {
			vulnerabilityResult.Epss = &types.Epss{
				Percentile: vulnerability.Epss.Percentile,
				Score:      vulnerability.Epss.Score,
			}
		}

		result = append(result, vulnerabilityResult)
	}

	return result
}
