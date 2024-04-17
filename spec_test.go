package skill

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseSpec_HandlesMultipleSkillDocs(t *testing.T) {
	file, _ := os.ReadFile("./test_data/skill.yaml")

	spec, err := ParseSpec(file)
	if err != nil {
		panic(err)
	}

	noFixablePolicy := spec["atomist/no-fixable-packages-goal"]

	assert.Equal(t, "no-fixable-packages-goal", noFixablePolicy.Name)
	assert.Equal(t, "Report on vulnerabilities that can be fixed by upgrading", noFixablePolicy.Description)
	assert.Equal(t, 30, getParameter("age", noFixablePolicy.YamlParameters).DefaultValue)

	badCvesPolicy := spec["docker/bad-cves-goal"]

	assert.Equal(t, "bad-cves-goal", badCvesPolicy.Name)
	assert.Equal(t, "Report on presence of high-profile CVEs", badCvesPolicy.Description)
	assert.Equal(t, []interface{}{"CVE-2023-38545", "CVE-2023-44487", "CVE-2014-0160", "CVE-2021-44228", "CVE-2024-3094"}, getParameter("cves", badCvesPolicy.YamlParameters).DefaultValues)
}

func getParameter(name string, parameters []map[string]ParameterSpecs) ParameterSpecs {
	for _, p := range parameters {
		for _, v := range p {
			if v.Name == name {
				return v
			}
		}
	}
	return ParameterSpecs{}
}
