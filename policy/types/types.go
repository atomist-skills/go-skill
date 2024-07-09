package types

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/openvex/go-vex/pkg/vex"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
)

const (
	BuildKitMaxMode = "buildkit_max_mode"
	BuildKitMinMode = "buildkit_min_mode"
)

type SBOM struct {
	Source          Source                  `json:"source"`
	Attestations    []dsse.Envelope         `json:"attestations"`
	Artifacts       []Package               `json:"artifacts"`
	Vulnerabilities []VulnerabilitiesByPurl `json:"vulnerabilities,omitempty"`
	VexDocuments    []vex.VEX               `json:"vex_statements,omitempty"`
	Secrets         []Secret                `json:"secrets,omitempty"`
	Descriptor      Descriptor              `json:"descriptor"`
}

type Source struct {
	Type       string            `json:"type"`
	Image      *ImageSource      `json:"image,omitempty"`
	FileSystem *FileSystemSource `json:"file_system,omitempty"`
	BaseImages []BaseImageMatch  `json:"base_images,omitempty"`
	Provenance *Provenance       `json:"provenance,omitempty"`
}

type Package struct {
	Type          string     `json:"type"`
	Namespace     string     `json:"namespace,omitempty"`
	Name          string     `json:"name"`
	Version       string     `json:"version"`
	Purl          string     `json:"purl"`
	Author        string     `json:"author,omitempty"`
	Description   string     `json:"description,omitempty"`
	Licenses      []string   `json:"licenses,omitempty"`
	Url           string     `json:"url,omitempty"`
	Size          int        `json:"size,omitempty"`
	InstalledSize int        `json:"installed_size,omitempty"`
	Locations     []Location `json:"locations"`
	Files         []Location `json:"files,omitempty"`
	Parent        string     `json:"parent,omitempty"`
}
type Secret struct {
	Source   SecretSource    `json:"source"`
	Findings []SecretFinding `json:"findings"`
}
type SecretSource struct {
	Type     string    `json:"type"`
	Location *Location `json:"location,omitempty"`
}

type SecretFinding struct {
	RuleID    string `json:"rule_id"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	Severity  string `json:"severity"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Match     string `json:"match"`
}
type Descriptor struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	SbomVersion string `json:"sbom_version"`
}

type Location struct {
	Path    string `json:"path,omitempty"`
	Ordinal int    `json:"ordinal,omitempty"`
	Digest  string `json:"digest,omitempty"`
	DiffID  string `json:"diff_id,omitempty"`
}

type ImageSource struct {
	Name        string         `json:"name"`
	Digest      string         `json:"digest"`
	Tags        *[]string      `json:"tags,omitempty"`
	Manifest    *v1.Manifest   `json:"manifest,omitempty"`
	Config      *v1.ConfigFile `json:"config,omitempty"`
	RawManifest string         `json:"raw_manifest"`
	RawConfig   string         `json:"raw_config"`
	Distro      Distro         `json:"distro"`
	Platform    Platform       `json:"platform"`
	Size        int64          `json:"size"`
	Details     *BaseImage     `json:"details,omitempty"`
}
type Distro struct {
	OsName    string `json:"os_name,omitempty"`
	OsVersion string `json:"os_version,omitempty"`
	OsDistro  string `json:"os_distro,omitempty"`
}
type Platform struct {
	Os           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
}
type FileSystemSource struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type BaseImageMatch struct {
	DiffIds []string    `graphql:"matches" json:"diff_ids,omitempty"`
	Images  []BaseImage `graphql:"images" json:"images,omitempty"`
}

type Provenance struct {
	SourceMap *SourceMap           `json:"source_map,omitempty"`
	VCS       *VCS                 `json:"vcs,omitempty"`
	BaseImage *ProvenanceBaseImage `json:"base_image,omitempty"`
	Stream    *Stream              `json:"stream,omitempty"`
	Mode      string               `json:"mode,omitempty"`
}

type SourceMap struct {
	Instructions []InstructionSourceMap `json:"instructions,omitempty"`
	Dockerfile   string                 `json:"dockerfile,omitempty"`
	Sha          string                 `json:"sha,omitempty"`
}

type VCS struct {
	Revision string `json:"revision,omitempty"`
	Source   string `json:"source,omitempty"`
}

type ProvenanceBaseImage struct {
	Name     string `json:"name,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Digest   string `json:"digest,omitempty"`
	Platform *v1.Platform `json:"platform,omitempty"`
}

type Stream struct {
	Name string `json:"name,omitempty"`
}

type InstructionSourceMap struct {
	Digests     []string `json:"digests,omitempty"`
	DiffIDs     []string `json:"diff_ids,omitempty"`
	Instruction string   `json:"instruction,omitempty"`
	Source      string   `json:"source,omitempty"`
	Path        string   `json:"path,omitempty"`
	StartLine   int      `json:"start_line,omitempty"`
	StartColumn int      `json:"start_column,omitempty"`
	EndLine     int      `json:"end_line,omitempty"`
	EndColumn   int      `json:"end_column,omitempty"`
}
