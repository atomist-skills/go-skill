package evaluators

import (
	"context"

	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/query"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

type EvaluatorFlags uint8

const (
	EVAL_SKIP_LOCAL EvaluatorFlags = 1 << iota
	OPT_DIGEST
)

type GoalEvaluator interface {
	GetFlags() EvaluatorFlags
	EvaluateGoal(ctx context.Context, req skill.RequestContext, commonData query.CommonSubscriptionQueryResult, subscriptionResults [][]edn.RawMessage) ([]goals.GoalEvaluationQueryResult, error)
}
