package skill

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

type SkillDoc struct {
	Skill SkillSpec `yaml:"skill" json:"skill"`
}

type SkillSpec struct {
	APIVersion            string                      `yaml:"apiVersion" json:"apiVersion"`
	Author                string                      `yaml:"author" json:"author"`
	CapabilitiesSpec      CapabilitiesSpec            `yaml:"capabilititesSpec" json:"capabilitiesSpec"`
	Commands              []Commands                  `yaml:"commands" json:"commands"`
	CreatedAt             time.Time                   `yaml:"createdAt" json:"createdAt"`
	DatalogSubscriptions  []DatalogSubscriptions      `yaml:"datalogSubscriptions" json:"datalogSubscriptions"`
	Derived               bool                        `yaml:"derived" json:"derived"`
	Description           string                      `yaml:"description" json:"description"`
	DispatchStyle         string                      `yaml:"dispatchStyle" json:"dispatchStyle"`
	DisplayName           string                      `yaml:"displayName" json:"displayName"`
	HomepageURL           string                      `yaml:"homepageUrl" json:"homepageUrl"`
	IconURL               string                      `yaml:"iconUrl" json:"iconUrl"`
	InCatalog             bool                        `yaml:"inCatalog" json:"inCatalog"`
	Ingesters             []string                    `yaml:"ingesters" json:"ingesters"`
	Integration           bool                        `yaml:"integration" json:"integration"`
	License               string                      `yaml:"license" json:"license"`
	LongDescription       string                      `yaml:"longDescription" json:"longDescription"`
	Maturities            []string                    `yaml:"maturities" json:"maturities"`
	MaxConfigurations     int                         `yaml:"maxConfigurations" json:"maxConfigurations"`
	Name                  string                      `yaml:"name" json:"name"`
	Namespace             string                      `yaml:"namespace" json:"namespace"`
	Owner                 bool                        `yaml:"owner" json:"owner"`
	ParameterSpecs        []ParameterSpecs            `json:"parameterSpecs"`
	YamlParameters        []map[string]ParameterSpecs `yaml:"parameters"`
	Platform              string                      `yaml:"platform" json:"platform"`
	PublishedAt           time.Time                   `yaml:"publishedAt" json:"publishedAt"`
	Readme                string                      `yaml:"readme" json:"readme"`
	ResourceProviderSpecs []ResourceProviderSpecs     `yaml:"resourceProviderSpecs" json:"resourceProviderSpecs"`
	Rules                 *[]string                   `yaml:"rules,omitempty" json:"rules,omitempty"`
	Schemata              []Schemata                  `yaml:"schemata" json:"schemata"`
	Subscriptions         []string                    `yaml:"subscriptions" json:"subscriptions"`
	Target                *Target                     `yaml:"target,omitempty" json:"target,omitempty"`
	Technologies          []string                    `yaml:"technologies" json:"technologies"`
	Version               string                      `yaml:"version" json:"version"`
	VideoURL              string                      `yaml:"videoUrl" json:"videoUrl"`
}

type Env struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}
type Limit struct {
	CPU    float32 `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory float32 `yaml:"memory,omitempty" json:"memory,omitempty"`
}
type Request struct {
	CPU    float32 `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory float32 `yaml:"memory,omitempty" json:"memory,omitempty"`
}
type Resources struct {
	Limit   *Limit   `yaml:"limit,omitempty" json:"limit,omitempty"`
	Request *Request `yaml:"request,omitempty" json:"request,omitempty"`
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
type Commands struct {
	Description string `yaml:"description" json:"description"`
	DisplayName string `yaml:"displayName" json:"displayName"`
	Name        string `yaml:"name" json:"name"`
	Pattern     string `yaml:"pattern" json:"pattern"`
}
type DatalogSubscriptions struct {
	Limit int    `yaml:"limit" json:"limit"`
	Name  string `yaml:"name" json:"name"`
	Query string `yaml:"query" json:"query"`
}
type ParameterSpecs struct {
	Description   string        `yaml:"description" json:"description"`
	DisplayName   string        `yaml:"displayName" json:"displayName"`
	Name          string        `yaml:"name" json:"name"`
	Required      bool          `yaml:"required" json:"required"`
	Visibility    string        `yaml:"visibility" json:"visibility"`
	DefaultValue  interface{}   `yaml:"defaultValue" json:"defaultValue"`
	DefaultValues []interface{} `yaml:"defaultValues" json:"defaultValues"`
	Type          string        `yaml:"type" json:"type"`
	Options       []OptionSpecs `yaml:"options" json:"options"`
}
type OptionSpecs struct {
	Description string `yaml:"description" json:"description"`
	Text        string `yaml:"text" json:"text"`
	Value       string `yaml:"value" json:"value"`
}
type ResourceProviderSpecs struct {
	Description string `yaml:"description" json:"description"`
	DisplayName string `yaml:"displayName" json:"displayName"`
	MaxAllowed  int    `yaml:"maxAllowed" json:"maxAllowed"`
	MinRequired int    `yaml:"minRequired" json:"minRequired"`
	Name        string `yaml:"name" json:"name"`
	TypeName    string `yaml:"typeName" json:"typeName"`
}
type Schemata struct {
	Name   string `yaml:"name" json:"name"`
	Schema string `yaml:"schema" json:"schema"`
}
type Headers struct {
	AdditionalProp1 string `yaml:"additionalProp1" json:"additionalProp1"`
	AdditionalProp2 string `yaml:"additionalProp2" json:"additionalProp2"`
	AdditionalProp3 string `yaml:"additionalProp3" json:"additionalProp3"`
}
type Target struct {
	Headers    Headers `yaml:"headers" json:"headers"`
	SigningKey string  `yaml:"signingKey" json:"signingKey"`
	Type       string  `yaml:"type" json:"type"`
	URL        string  `yaml:"url" json:"url"`
}

func ParseSpec(data []byte) (map[string]SkillSpec, error) {
	result := map[string]SkillSpec{}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var spec SkillDoc
		err := decoder.Decode(&spec)

		// break the loop in case of EOF
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		if spec.Skill.Name != "" {
			namespace := "atomist"
			if spec.Skill.Namespace != "" {
				namespace = spec.Skill.Namespace
			}

			fullName := fmt.Sprintf("%s/%s", namespace, spec.Skill.Name)

			parameterSpecs := []ParameterSpecs{}
			for _, p := range spec.Skill.YamlParameters {
				for t, v := range p {
					v.Type = t
					parameterSpecs = append(parameterSpecs, v)
				}
			}

			spec.Skill.ParameterSpecs = parameterSpecs

			result[fullName] = spec.Skill
		}
	}

	return result, nil
}
