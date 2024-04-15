package skill

import "time"

type SkillSpec struct {
	APIVersion            string                  `json:"apiVersion"`
	Artifacts             []Artifacts             `json:"artifacts"`
	Author                string                  `json:"author"`
	CapabilitiesSpec      CapabilitiesSpec        `json:"capabilitiesSpec"`
	Categories            []Categories            `json:"categories"`
	Commands              []Commands              `json:"commands"`
	CreatedAt             time.Time               `json:"createdAt"`
	DatalogSubscriptions  []DatalogSubscriptions  `json:"datalogSubscriptions"`
	Derived               bool                    `json:"derived"`
	Description           string                  `json:"description"`
	DispatchStyle         string                  `json:"dispatchStyle"`
	DisplayName           string                  `json:"displayName"`
	HomepageURL           string                  `json:"homepageUrl"`
	IconURL               string                  `json:"iconUrl"`
	InCatalog             bool                    `json:"inCatalog"`
	Ingesters             []string                `json:"ingesters"`
	Integration           bool                    `json:"integration"`
	License               string                  `json:"license"`
	LongDescription       string                  `json:"longDescription"`
	Maturities            []string                `json:"maturities"`
	MaxConfigurations     int                     `json:"maxConfigurations"`
	Name                  string                  `json:"name"`
	Namespace             string                  `json:"namespace"`
	Owner                 bool                    `json:"owner"`
	ParameterSpecs        []ParameterSpecs        `json:"parameterSpecs"`
	Platform              string                  `json:"platform"`
	PublishedAt           time.Time               `json:"publishedAt"`
	Readme                string                  `json:"readme"`
	ResourceProviderSpecs []ResourceProviderSpecs `json:"resourceProviderSpecs"`
	Rules                 *[]string               `json:"rules,omitempty"`
	Schemata              []Schemata              `json:"schemata"`
	Subscriptions         []string                `json:"subscriptions"`
	Target                *Target                 `json:"target,omitempty"`
	Technologies          []string                `json:"technologies"`
	Version               string                  `json:"version"`
	VideoURL              string                  `json:"videoUrl"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Limit struct {
	CPU    float32 `json:"cpu,omitempty"`
	Memory float32 `json:"memory,omitempty"`
}
type Request struct {
	CPU    float32 `json:"cpu,omitempty"`
	Memory float32 `json:"memory,omitempty"`
}
type Resources struct {
	Limit   *Limit   `json:"limit,omitempty"`
	Request *Request `json:"request,omitempty"`
}
type Artifacts struct {
	Name       string    `json:"name"`
	Args       []string  `json:"args"`
	Command    []string  `json:"command"`
	Env        []Env     `json:"env"`
	Image      string    `json:"image"`
	Resources  Resources `json:"resources"`
	Type       string    `json:"type"`
	WorkingDir string    `json:"workingDir"`
}
type Declares struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}
type Provides struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}
type Catalog struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
}
type Configured struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
}
type Other struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
}
type Owned struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
}
type Providers struct {
	Catalog    []Catalog    `yaml:"catalog" json:"catalog"`
	Configured []Configured `yaml:"configured" json:"configured"`
	Other      []Other      `yaml:"other" json:"other"`
	Owned      []Owned      `yaml:"owned" json:"owned"`
}
type Requires struct {
	Description string    `yaml:"description" json:"description"`
	DisplayName string    `yaml:"description" json:"displayName"`
	MaxAllowed  *int      `yaml:"maxAllowed" json:"maxAllowed"`
	MinRequired *int      `yaml:"minRequired" json:"minRequired"`
	Name        string    `yaml:"name" json:"name"`
	Namespace   string    `yaml:"namespace" json:"namespace"`
	Providers   Providers `yaml:"providers" json:"providers"`
	Scopes      []string  `yaml:"scopes" json:"scopes"`
	Usage       string    `yaml:"usage" json:"usage"`
}
type CapabilitiesSpec struct {
	Declares []Declares `yaml:"declares,omitempty" json:"declares,omitempty"`
	Provides []Provides `yaml:"provides,omitempty" json:"provides,omitempty"`
	Requires []Requires `yaml:"requires,omitempty" json:"requires,omitempty"`
}
type Categories struct {
	Key       string `json:"key"`
	SortOrder int    `json:"sortOrder"`
	Text      string `json:"text"`
}
type Commands struct {
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
}
type DatalogSubscriptions struct {
	Limit int    `yaml:"limit" json:"limit"`
	Name  string `yaml:"name" json:"name"`
	Query string `yaml:"query" json:"query"`
}
type ParameterSpecs struct {
	Description  string        `json:"description"`
	DisplayName  string        `json:"displayName"`
	Name         string        `json:"name"`
	Required     bool          `json:"required"`
	Visibility   string        `json:"visibility"`
	DefaultValue interface{}   `json:"defaultValue"`
	Type         string        `json:"type"`
	Options      []OptionSpecs `json:"options"`
}
type OptionSpecs struct {
	Description string `json:"description"`
	Text        string `json:"text"`
	Value       string `json:"value"`
}
type ResourceProviderSpecs struct {
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
	MaxAllowed  int    `json:"maxAllowed"`
	MinRequired int    `json:"minRequired"`
	Name        string `json:"name"`
	TypeName    string `json:"typeName"`
}
type Schemata struct {
	Name   string `yaml:"name" json:"name"`
	Schema string `yaml:"schema" json:"schema"`
}
type Headers struct {
	AdditionalProp1 string `json:"additionalProp1"`
	AdditionalProp2 string `json:"additionalProp2"`
	AdditionalProp3 string `json:"additionalProp3"`
}
type Target struct {
	Headers    Headers `json:"headers"`
	SigningKey string  `json:"signingKey"`
	Type       string  `json:"type"`
	URL        string  `json:"url"`
}
