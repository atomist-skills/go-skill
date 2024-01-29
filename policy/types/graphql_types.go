package types

import "github.com/openvex/go-vex/pkg/vex"

type BaseImage struct {
	CreatedAt  string              `graphql:"createdAt" json:"created_at,omitempty"`
	Digest     string              `graphql:"digest" json:"digest,omitempty"`
	Repository BaseImageRepository `graphql:"repository" json:"repository"`
	Tags       []struct {
		Current   bool   `graphql:"current" json:"current"`
		Name      string `graphql:"name" json:"name,omitempty"`
		Supported bool   `graphql:"supported" json:"supported"`
	} `graphql:"tags" json:"tags,omitempty"`
	DockerFile struct {
		Commit struct {
			Repository struct {
				Org  string `graphql:"orgName" json:"org,omitempty"`
				Repo string `graphql:"repoName" json:"repo,omitempty"`
			} `graphql:"repository" json:"repository,omitempty"`
			Sha string `graphql:"sha" json:"sha,omitempty"`
		} `json:"commit,omitempty"`
		Path string `graphql:"path" json:"path,omitempty"`
	} `graphql:"dockerFile" json:"docker_file,omitempty"`
	PackageCount        int                  `graphql:"packageCount" json:"package_count,omitempty"`
	VulnerabilityReport *VulnerabilityReport `graphql:"vulnerabilityReport" json:"vulnerability_report"`
	Platform            struct {
		Arch    string `graphql:"architecture"`
		OS      string `graphql:"os"`
		Variant string `graphql:"variant"`
	} `graphql:"platform"`
}
type VulnerabilitiesByPurl struct {
	Purl            string          `graphql:"purl" json:"purl,omitempty"`
	Vulnerabilities []Vulnerability `graphql:"vulnerabilities" json:"vulnerabilities,omitempty"`
}

type Vulnerability struct {
	Source          string `graphql:"source" json:"source,omitempty"`
	SourceId        string `graphql:"sourceId" json:"source_id,omitempty"`
	Description     string `graphql:"description" json:"description,omitempty"`
	VulnerableRange string `graphql:"vulnerableRange" json:"vulnerable_range,omitempty"`
	FixedBy         string `graphql:"fixedBy" json:"fixed_by,omitempty"`
	Url             string `graphql:"url" json:"url,omitempty"`
	Cvss            struct {
		Score    float32 `graphql:"score" json:"score,omitempty"`
		Severity string  `graphql:"severity" json:"severity,omitempty"`
		Vector   string  `graphql:"vector" json:"vector,omitempty"`
		Version  string  `graphql:"version" json:"version,omitempty"`
	} `graphql:"cvss" json:"cvss,omitempty"`
	Cwes          []Cwe     `graphql:"cwes" json:"cwes,omitempty"`
	VexStatements []vex.VEX `graphql:"-" json:"vex_statements,omitempty"`
}

type BaseImageRepository struct {
	Badge         string   `graphql:"badge" json:"badge,omitempty"`
	Host          string   `graphql:"hostName" json:"host,omitempty"`
	Repo          string   `graphql:"repoName" json:"repo,omitempty"`
	SupportedTags []string `graphql:"supportedTags" json:"supported_tags,omitempty"`
	PreferredTags []string `graphql:"preferredTags" json:"preferred_tags,omitempty"`
}

type VulnerabilityReport struct {
	Critical    int `graphql:"critical" json:"critical,omitempty"`
	High        int `graphql:"high" json:"high,omitempty"`
	Medium      int `graphql:"medium" json:"medium,omitempty"`
	Low         int `graphql:"low" json:"low,omitempty"`
	Unspecified int `graphql:"unspecified" json:"unspecified,omitempty"`
	Total       int `graphql:"total" json:"total,omitempty"`
}
type Cwe struct {
	CweId string `graphql:"cweId" json:"cwe_id,omitempty"`
	Name  string `graphql:"description" json:"name,omitempty"`
}
