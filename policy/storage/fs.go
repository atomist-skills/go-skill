package storage

import (
	"context"
	"os"

	"github.com/atomist-skills/go-skill/policy/goals"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

type FsStorage struct {
	path string
}

func NewFsStorage(ctx context.Context) (EvaluationStorage, error) {
	return &FsStorage{
		path: os.TempDir(),
	}, nil
}

func (f *FsStorage) Store(ctx context.Context, results []goals.GoalEvaluationQueryResult, storageId string, log skill.Logger) error {
	log.Infof("Storing %d results", len(results))

	content, err := edn.Marshal(results)
	if err != nil {
		return err
	}
	log.Infof("Content to store: %s", string(content))

	file, err := os.Create(f.path + "/results.edn")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(string(content))
	if err != nil {
		return err
	}

	return nil
}
