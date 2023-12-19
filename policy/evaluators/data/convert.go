package data

import (
	"time"

	"github.com/atomist-skills/go-skill/policy/graphql"
)

func getPurlsFromPackages(packages []MetadataPackage) []string {
	purls := []string{}
	for _, pkg := range packages {
		purls = append(purls, pkg.Purl)
	}
	return purls
}

func getPackagesByPurl(packages []graphql.VulnerabilitiesByPackage) map[string]graphql.VulnerabilitiesByPackage {
	result := map[string]graphql.VulnerabilitiesByPackage{}
	for _, pkg := range packages {
		result[pkg.Purl] = pkg
	}

	return result
}

func convertGraphqlToPackages(imagePackages graphql.ImagePackagesByDigest) ([]Package, error) {
	nonEmptyHistories := []graphql.ImageHistory{}
	for _, history := range imagePackages.ImageHistories {
		if !history.EmptyLayer {
			nonEmptyHistories = append(nonEmptyHistories, history)
		}
	}

	pkgs := []Package{}
	for _, p := range imagePackages.ImagePackages.Packages {
		locations := []PackageLocation{}
		for _, location := range p.PackageLocations {
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

			locations = append(locations, PackageLocation{
				LayerOrdinal: historyOrdinal,
				Path:         location.Path,
			})
		}

		var namespace string
		if p.Package.Namespace == nil {
			namespace = ""
		} else {
			namespace = *p.Package.Namespace
		}

		vulnerabilities, err := convertToVulnerabilities(p.Package.Vulnerabilities)
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, Package{
			Purl:            p.Package.Purl,
			Licenses:        p.Package.Licenses,
			Name:            p.Package.Name,
			Namespace:       namespace,
			Version:         p.Package.Version,
			Type:            p.Package.Type,
			Locations:       locations,
			Vulnerabilities: vulnerabilities,
		})
	}

	return pkgs, nil
}

func convertMetadataPackagesToPackages(metadataPackages []MetadataPackage, vulnerabilitiesByPackage []graphql.VulnerabilitiesByPackage) ([]Package, error) {
	pkgsByPurl := getPackagesByPurl(vulnerabilitiesByPackage)

	packages := []Package{}
	for _, mPkg := range metadataPackages {
		pkg := pkgsByPurl[mPkg.Purl]
		vulnerabilities, err := convertToVulnerabilities(pkg.Vulnerabilities)
		if err != nil {
			return nil, err
		}

		packages = append(packages, Package{
			Licenses:        mPkg.Licenses,
			Name:            mPkg.Name,
			Namespace:       mPkg.Namespace,
			Version:         mPkg.Version,
			Purl:            mPkg.Purl,
			Type:            mPkg.Type,
			Vulnerabilities: vulnerabilities,
		})
	}

	return packages, nil
}

func convertToVulnerabilities(vulnerabilities []graphql.Vulnerability) ([]Vulnerability, error) {
	result := []Vulnerability{}

	for _, vulnerability := range vulnerabilities {
		publishedAt, err := time.Parse(time.RFC3339, vulnerability.PublishedAt)
		if err != nil {
			return nil, err
		}

		updatedAt, err := time.Parse(time.RFC3339, vulnerability.UpdatedAt)
		if err != nil {
			return nil, err
		}

		vulnerabilityResult := Vulnerability{
			Cvss:            Cvss{},
			PublishedAt:     publishedAt,
			Source:          vulnerability.Source,
			SourceID:        vulnerability.SourceID,
			UpdatedAt:       updatedAt,
			VulnerableRange: vulnerability.VulnerableRange,
		}

		if vulnerability.Cvss.Score != nil {
			vulnerabilityResult.Cvss.Score = *vulnerability.Cvss.Score
		}

		if vulnerability.Cvss.Severity != nil {
			vulnerabilityResult.Cvss.Severity = *vulnerability.Cvss.Severity
		}

		if vulnerability.URL != nil {
			vulnerabilityResult.URL = *vulnerability.URL
		}

		if vulnerability.FixedBy != nil {
			vulnerabilityResult.FixedBy = *vulnerability.FixedBy
		}

		result = append(result, vulnerabilityResult)
	}

	return result, nil
}
