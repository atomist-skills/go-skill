package graphql

type GqlContext struct {
	TeamId       string `json:"teamId"`
	Organization string `json:"organization"`
}

type ImagePlatform struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

type BaseImage struct {
	Digest     string     `json:"digest" edn:"digest"`
	Repository Repository `json:"repository" edn:"repository"`
	Tags       []Tag      `json:"tags" edn:"tags"`
}

type Repository struct {
	HostName string `json:"hostName" edn:"hostName"`
	RepoName string `json:"repoName" edn:"repoName"`
}

type Tag struct {
	Name    string `json:"name" edn:"name"`
	Current bool   `json:"current" edn:"current"`
}

type ImagePackages struct {
	Packages []Packages `json:"packages" edn:"packages"`
}

type ImageHistory struct {
	EmptyLayer bool `json:"emptyLayer" edn:"emptyLayer"`
	Ordinal    int  `json:"ordinal" edn:"ordinal"`
}

type ImageLayers struct {
	Layers []ImageLayer `json:"layers" edn:"layers"`
}

type ImageLayer struct {
	DiffId  string `json:"diffId" edn:"diffId"`
	Ordinal int    `json:"ordinal" edn:"ordinal"`
}

type Packages struct {
	Package          PackageWithLicenses `json:"package" edn:"package"`
	PackageLocations []PackageLocation   `json:"locations" edn:"locations"`
}

type PackageWithLicenses struct {
	Licenses        []string        `json:"licenses" edn:"licenses"`
	Name            string          `json:"name" edn:"name"`
	Namespace       *string         `json:"namespace" edn:"namespace"`
	Version         string          `json:"version" edn:"version"`
	Purl            string          `json:"purl" edn:"purl"`
	Type            string          `json:"type" edn:"type"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities" edn:"vulnerabilities"`
}

type PackageLocation struct {
	DiffId string `json:"diffId" edn:"diffId"`
	Path   string `json:"path" edn:"path"`
}

type Vulnerability struct {
	Cvss            Cvss    `json:"cvss"`
	FixedBy         *string `json:"fixedBy"`
	PublishedAt     string  `json:"publishedAt"`
	Source          string  `json:"source"`
	SourceID        string  `json:"sourceId"`
	UpdatedAt       string  `json:"updatedAt"`
	URL             *string `json:"url"`
	VulnerableRange string  `json:"vulnerableRange"`
}

type Cvss struct {
	Severity *string  `json:"severity"`
	Score    *float32 `json:"score"`
}
