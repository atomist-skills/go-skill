package jynx

type (
	ImagePackagesByDigestResponse struct {
		ImagePackagesByDigest *ImagePackagesByDigest `json:"imagePackagesByDigest" edn:"imagePackagesByDigest"`
	}

	ImagePackagesByDigest struct {
		Digest         string         `json:"digest" edn:"digest"`
		ImagePackages  ImagePackages  `json:"imagePackages" edn:"imagePackages"`
		ImageHistories []ImageHistory `json:"imageHistories" edn:"imageHistories"`
		ImageLayers    ImageLayers    `json:"imageLayers" edn:"imageLayers"`
	}

	ImagePackages struct {
		Packages []Packages `json:"packages" edn:"packages"`
	}

	ImageHistory struct {
		EmptyLayer bool `json:"emptyLayer" edn:"emptyLayer"`
		Ordinal    int  `json:"ordinal" edn:"ordinal"`
	}

	ImageLayers struct {
		Layers []ImageLayer `json:"layers" edn:"layers"`
	}

	ImageLayer struct {
		DiffId  string `json:"diffId" edn:"diffId"`
		Ordinal int    `json:"ordinal" edn:"ordinal"`
	}

	Packages struct {
		Package   Package           `json:"package" edn:"package"`
		Locations []PackageLocation `json:"locations" edn:"locations"`
	}

	Package struct {
		Licenses        []string        `json:"licenses" edn:"licenses"`
		Name            string          `json:"name" edn:"name"`
		Namespace       *string         `json:"namespace" edn:"namespace"`
		Version         string          `json:"version" edn:"version"`
		Purl            string          `json:"purl" edn:"purl"`
		Type            string          `json:"type" edn:"type"`
		Vulnerabilities []Vulnerability `json:"vulnerabilities" edn:"vulnerabilities"`
	}

	PackageLocation struct {
		DiffId string `json:"diffId" edn:"diffId"`
		Path   string `json:"path" edn:"path"`
	}

	Vulnerability struct {
		Cvss            Cvss    `json:"cvss" edn:"cvss"`
		Epss            *Epss   `json:"epss" edn:"epss"`
		FixedBy         *string `json:"fixedBy" edn:"fixedBy"`
		PublishedAt     string  `json:"publishedAt" edn:"publishedAt"`
		Source          string  `json:"source" edn:"source"`
		SourceID        string  `json:"sourceId" edn:"sourceId"`
		UpdatedAt       string  `json:"updatedAt" edn:"updatedAt"`
		URL             *string `json:"url" edn:"url"`
		VulnerableRange string  `json:"vulnerableRange" edn:"vulnerableRange"`
		CisaExploited   bool    `json:"cisaExploited" edn:"cisaExploited"`
	}

	Cvss struct {
		Severity *string  `json:"severity" edn:"severity"`
		Score    *float32 `json:"score" edn:"score"`
	}

	Epss struct {
		Percentile float32 `json:"percentile" edn:"percentile"`
		Score      float32 `json:"score" edn:"score"`
	}
)
