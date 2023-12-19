package data

import (
	"context"
	"time"

	"github.com/atomist-skills/go-skill/policy/graphql"
	"github.com/atomist-skills/go-skill/policy/query"
)

type Package struct {
	Licenses        []string
	Name            string
	Namespace       string
	Version         string
	Purl            string
	Type            string
	Locations       []PackageLocation
	Vulnerabilities []Vulnerability
}

type PackageLocation struct {
	LayerOrdinal int
	Path         string
}

type Vulnerability struct {
	Cvss            Cvss
	FixedBy         string
	PublishedAt     time.Time
	Source          string
	SourceID        string
	UpdatedAt       time.Time
	URL             string
	VulnerableRange string
}

type Cvss struct {
	Severity string
	Score    float32
}

type GetPackagesResult struct {
	AsyncQueryMade bool
	Result         []Package
}

type GetImageDetailsByDigestResult struct {
	AsyncQueryMade bool
	Result         *graphql.ImageDetailsByDigest
}

type DataSource interface {
	GetPackages(ctx context.Context, digest string) (*GetPackagesResult, error)
	GetImageDetailsByDigest(ctx context.Context, digest string, platform query.ImagePlatform) (*GetImageDetailsByDigestResult, error)
}
