package normalization

import (
	"fmt"
	"strings"

	anchorepackageurl "github.com/anchore/packageurl-go"
	"github.com/anchore/syft/syft/linux"
	"github.com/atomist-skills/go-skill/internal"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/openvex/go-vex/pkg/vex"
)

// NormalizeSBOM creates the canonical representation of our internal PURL model
// This has to be moved into the backend API layer some time in the future
func NormalizeSBOM(sb *types.SBOM) ([]string, map[string]string) {
	purls := make([]string, 0)
	purlMapping := make(map[string]string)

	var d *types.Distro
	if sb.Source.Type == "image" {
		d = &sb.Source.Image.Distro
	}
	for _, pkg := range sb.Artifacts {
		purl, upstreamPurl := NormalizePURL(pkg.Purl, d)
		purls = append(purls, purl)
		purlMapping[purl] = pkg.Purl
		if upstreamPurl != "" {
			purls = append(purls, upstreamPurl)
			purlMapping[upstreamPurl] = pkg.Purl
		}
	}
	return purls, purlMapping
}

func NormalizePURL(pkg string, d *types.Distro) (string, string) {
	var upstreamName, upstreamVersion string
	purl, _ := ToPackageUrl(pkg)
	if purl.Type == "deb" || purl.Type == "rpm" || purl.Type == "alpine" || purl.Type == "apk" {
		for _, q := range purl.Qualifiers {
			if q.Key == "distro" {
				seg := strings.Split(q.Value, "-")
				if len(seg) == 2 {
					_, qualifiers := OsQualifiers(&linux.Release{
						Name:    strings.Split(q.Value, "-")[0],
						Version: strings.Split(q.Value, "-")[1],
					})
					purl.Qualifiers = anchorepackageurl.QualifiersFromMap(qualifiers)
				} else {
					_, qualifiers := OsQualifiers(&linux.Release{
						Name: q.Value,
					})
					purl.Qualifiers = anchorepackageurl.QualifiersFromMap(qualifiers)
				}
			} else if q.Key == "upstream" {
				parts := strings.Split(q.Value, "@")
				upstreamName = parts[0]
				if len(parts) == 2 {
					upstreamVersion = parts[1]
				}
			}
		}
		// add the distro qualifiers when missing
		if d != nil {
			if _, ok := purl.Qualifiers.Map()["os_name"]; !ok {
				_, qualifiers := OsQualifiers(&linux.Release{
					Name:    d.OsName,
					Version: d.OsVersion,
				})
				purl.Qualifiers = anchorepackageurl.QualifiersFromMap(qualifiers)
			}
		}
	}
	// move subpath back into the name so that the backend can handle it
	if purl.Subpath != "" {
		if purl.Namespace == "" {
			purl.Namespace = purl.Name
		} else {
			purl.Namespace = fmt.Sprintf("%s/%s", purl.Namespace, purl.Name)
		}
		purl.Name = purl.Subpath
		purl.Subpath = ""
	}

	pkgPurl := purl.String()
	var upstreamPurl string
	if upstreamName != "" {
		purl.Name = upstreamName
		if upstreamVersion != "" {
			purl.Version = upstreamVersion
		}
		upstreamPurl = purl.String()
	}
	return pkgPurl, upstreamPurl
}

func ContainsPurl(purls []vex.Subcomponent, purl string) bool {
	p, _ := ToPackageUrl(purl)
	for _, pu := range purls {
		np, nsp := NormalizePURL(pu.ID, nil)
		npp, err := ToPackageUrl(np)
		if err != nil {
			continue
		}
		if npp.Type == p.Type && npp.Namespace == p.Namespace && npp.Name == p.Name && npp.Version == p.Version {
			return true
		}
		nspp, err := ToPackageUrl(nsp)
		if err != nil {
			continue
		}
		if nspp.Type == p.Type && nspp.Namespace == p.Namespace && nspp.Name == p.Name && nspp.Version == p.Version {
			return true
		}
	}
	return false
}

func ToPackageUrl(url string) (anchorepackageurl.PackageURL, error) {
	url = strings.TrimSuffix(url, "/")

	// once again, there's a strange round trip issue with purls coming out of syft.
	// this time it is the optional subpath loosing `.` and `..` segments which is really vague
	// in the spec at https://github.com/package-url/purl-spec/blob/master/PURL-SPECIFICATION.rst#rules-for-each-purl-component.
	parts := strings.Split(url, "#")
	var subpath string
	if len(parts) > 1 {
		subpath = parts[1]
	}
	purl, err := anchorepackageurl.FromString(url)
	purl.Subpath = subpath
	return purl, err
}

var NamespaceMapping = map[string]string{
	"oracle": "oraclelinux",
	"ol":     "oraclelinux",
	"amazon": "amazonlinux",
	"amzn":   "amazonlinux",
	"rhel":   "redhatlinux",
}

func OsQualifiers(release *linux.Release) (types.Distro, map[string]string) {
	qualifiers := make(map[string]string, 0)
	distro := types.Distro{}
	if release == nil {
		return distro, qualifiers
	}
	if release.ID != "" {
		distro.OsName = release.ID
	} else if release.Name != "" {
		distro.OsName = release.Name
	}
	if release.Version != "" {
		distro.OsVersion = release.Version
	} else if release.VersionID != "" {
		distro.OsVersion = release.VersionID
	}

	if v, ok := NamespaceMapping[distro.OsName]; ok {
		distro.OsName = v
	}

	if distro.OsVersion != "" {
		// alpine: with comma
		// amazonlinux: single digit
		// debian: single digit
		// oraclelinux: single digit
		// redhatlinux: single digit
		// centos: single digit
		// ubuntu: with comma
		version := strings.Split(distro.OsVersion, " ")[0]
		parts := strings.Split(version, ".")
		if distro.OsName == "alpine" || distro.OsName == "ubuntu" {
			distro.OsVersion = strings.Join(parts[0:2], ".")
		} else {
			distro.OsVersion = parts[0]
		}
	} else if distro.OsName == "debian" {
		distro.OsVersion = "unstable"
	}

	if release.VersionCodename != "" {
		distro.OsDistro = release.VersionCodename
	}

	// sometimes OsVersion contains _ eg for alpine:edge
	if strings.Contains(distro.OsVersion, "_") {
		distro.OsVersion = strings.Split(distro.OsVersion, "_")[0]
	}

	// special handling for wolfi images
	if distro.OsName == "wolfi" || distro.OsName == "chainguard" {
		qualifiers["os_name"] = distro.OsName
		qualifiers["os_version"] = "rolling"
		qualifiers["distro_name"] = distro.OsName
		qualifiers["distro_version"] = distro.OsVersion
	} else {
		qualifiers["os_name"] = distro.OsName
		qualifiers["os_version"] = distro.OsVersion
		if distro.OsDistro != "" {
			qualifiers["os_distro"] = distro.OsDistro
		}
	}
	return distro, qualifiers
}

func DenormalizeSBOM(vp *types.VulnerabilitiesByPurls, purlMapping map[string]string) {
	for i, p := range vp.VulnerabilitiesByPackage {
		vp.VulnerabilitiesByPackage[i].Purl = purlMapping[p.Purl]
	}
	// At the end we could end up with duplicate entries for packages
	purlsToCVEs := make(map[string][]types.Vulnerability)
	for _, p := range vp.VulnerabilitiesByPackage {
		if v, ok := purlsToCVEs[p.Purl]; ok {
			for _, pcve := range p.Vulnerabilities {
				if !internal.ContainsBy(v, func(cve types.Vulnerability) bool {
					return cve.SourceId == pcve.SourceId && cve.Source == pcve.Source
				}) {
					v = append(v, pcve)
					purlsToCVEs[p.Purl] = v
				}
			}
		} else {
			purlsToCVEs[p.Purl] = p.Vulnerabilities
		}
	}
	vp.VulnerabilitiesByPackage = make([]types.VulnerabilitiesByPurl, 0)
	for k, v := range purlsToCVEs {
		vp.VulnerabilitiesByPackage = append(vp.VulnerabilitiesByPackage, types.VulnerabilitiesByPurl{
			Purl:            k,
			Vulnerabilities: v,
		})
	}
}
