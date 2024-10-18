package policy

import (
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
)

const VulnerabilityChangeEvent = "VulnerabilityChangeEvent"

type Policy struct {
	CreateEvaluatorFunc func(map[string]interface{}, data.DataSource) (goals.GoalEvaluator, error)
	Spec                *skill.SkillSpec
	EventSubscriptions  []string
}
