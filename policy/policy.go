package policy

import (
	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/policy_handler"
)

const VulnerabilityChangeEvent = "VulnerabilityChangeEvent"

type Policy struct {
	SkillHandler        policy_handler.EventHandler
	CreateEvaluatorFunc func(map[string]interface{}, data.DataSource) (goals.GoalEvaluator, error)
	Spec                *skill.SkillSpec
	EventSubscriptions  []string
}
