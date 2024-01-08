package storage

import (
	"context"
	"os"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
)

type EvaluationStorage interface {
	Store(ctx context.Context, results []goals.GoalEvaluationQueryResult, storageId, environment string, log skill.Logger) error
}

// NewEvaluationStorage creates a new EvaluationStorage object based on the LOCAL_DEBUG environment variable.
func NewEvaluationStorage(ctx context.Context) (EvaluationStorage, error) {
	if os.Getenv("LOCAL_DEBUG") == "true" {
		return NewFsStorage(ctx)
	}

	return NewGcsStorage(ctx)
}
