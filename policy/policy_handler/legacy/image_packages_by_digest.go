package legacy

import (
	"context"
	"fmt"
	"strings"

	anchorepackageurl "github.com/anchore/packageurl-go"
	"github.com/anchore/syft/syft/linux"
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/internal"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/types"
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
		SourceId        string  `json:"sourceId"`
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
		purls, purlMapping = NormalizeSBOM(sb)
		req.Log.Debugf("Normalized purls: %+v", purls)
		req.Log.Debugf("Purl mapping: %+v", purlMapping)
	} else {
		req.Log.Debug("Using packages from metadata (legacy) for fetching vulnerabilities")
		for _, p := range legacyPkgs {
			purls = append(purls, p.Purl)
		}
	}

	var vulnsResponse VulnerabilitiesByPackageResponse
	_, err := ds.Query(ctx, vulnerabilitiesByPackageQueryName, vulnerabilitiesByPackageQuery, map[string]interface{}{
		"context":     data.GqlContext(req),
		"packageUrls": purls,
	}, &vulnsResponse)
	if err != nil {
		return ImagePackagesByDigestResponse{}, err
	}

	if sb != nil {
		req.Log.Debug("Denormalizing purls after fetching vulnerabilities")
		DenormalizeSBOM(&vulnsResponse, purlMapping)
	}

	vulns := map[string][]Vulnerability{}
	for _, v := range vulnsResponse.VulnerabilitiesByPackage {
		vulns[v.Purl] = v.Vulnerabilities
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
					Vulnerabilities: vulns[a.Purl],
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
					Vulnerabilities: vulns[a.Purl],
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

const (
	SourceTypeImage = "image"
)

// NormalizeSBOM creates the canonical representation of our internal PURL model
// This has to be moved into the backend API layer some time in the future
func NormalizeSBOM(sb *types.SBOM) ([]string, map[string]string) {
	purls := make([]string, 0)
	purlMapping := make(map[string]string)

	var d *types.Distro
	if sb.Source.Type == SourceTypeImage {
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

func DenormalizeSBOM(vp *VulnerabilitiesByPackageResponse, purlMapping map[string]string) {
	for i, p := range vp.VulnerabilitiesByPackage {
		vp.VulnerabilitiesByPackage[i].Purl = purlMapping[p.Purl]
	}
	// At the end we could end up with duplicate entries for packages
	purlsToCVEs := make(map[string][]Vulnerability)
	for _, p := range vp.VulnerabilitiesByPackage {
		if v, ok := purlsToCVEs[p.Purl]; ok {
			for _, pcve := range p.Vulnerabilities {
				if !internal.ContainsBy(v, func(cve Vulnerability) bool {
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
	vp.VulnerabilitiesByPackage = make([]VulnerabilitiesByPackage, 0)
	for k, v := range purlsToCVEs {
		vp.VulnerabilitiesByPackage = append(vp.VulnerabilitiesByPackage, VulnerabilitiesByPackage{
			Purl:            k,
			Vulnerabilities: v,
		})
	}
}

var NamespaceMapping = map[string]string{
	"oracle": "oraclelinux",
	"ol":     "oraclelinux",
	"amazon": "amazonlinux",
	"amzn":   "amazonlinux",
	"rhel":   "redhatlinux",
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
