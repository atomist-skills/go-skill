package transact

import (
	"context"
	"fmt"
	"time"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/storage"
)

type PreviousResult struct {
	StorageId  string
	ConfigHash string
}

func TransactPolicyResult(
	ctx context.Context,
	evalCtx goals.GoalEvaluationContext,
	configuration skill.Configuration,
	digest string,
	previousResult *PreviousResult,
	evaluationTs time.Time,
	goalResults []goals.GoalEvaluationQueryResult,
	tx int64,
	newTransaction func() skill.Transaction,
) (*goals.GoalEvaluationResultEntity, error) {
	var previousConfigHash, previousStorageId string
	if previousResult == nil {
		previousConfigHash = "n/a"
		previousStorageId = "n/a"
	} else {
		previousConfigHash = previousResult.ConfigHash
		previousStorageId = previousResult.StorageId
	}

	if goalResults == nil {
		evalCtx.Log.Infof("returned no data for digest %s", digest)
	}

	es, err := storage.NewEvaluationStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to create evaluation storage: %s", err.Error())
	}

	configDiffer, configHash, err := goals.GoalConfigsDiffer(evalCtx.Log, configuration, digest, previousConfigHash)
	if err != nil {
		evalCtx.Log.Errorf("Failed to check if config hash changed for digest: %s", digest, err)
		evalCtx.Log.Warnf("Will continue with the evaluation nonetheless")
		configDiffer = true
	}

	differ, storageId, err := goals.GoalResultsDiffer(evalCtx.Log, goalResults, digest, previousStorageId)
	if err != nil {
		evalCtx.Log.Errorf("Failed to check if goal results changed for digest: %s", digest, err)
		evalCtx.Log.Warnf("Will continue with the evaluation nonetheless")
		differ = true
	}

	if differ && goalResults != nil {
		if err := es.Store(ctx, goalResults, storageId, evalCtx.Log); err != nil {
			return nil, fmt.Errorf("Failed to store evaluation results for digest %s: %s", digest, err.Error())
		}
	}

	var resultEntity *goals.GoalEvaluationResultEntity
	if differ || configDiffer {
		shouldRetract := previousStorageId != "no-data" && previousStorageId != "n/a" && storageId == "no-data"
		entity := goals.CreateEntitiesFromResults(goalResults, evalCtx.Goal.Definition, evalCtx.Goal.Configuration, digest, storageId, configHash, evaluationTs, tx, shouldRetract)
		resultEntity = &entity
	}

	if resultEntity != nil {
		err = newTransaction().AddEntities(*resultEntity).Transact()
		if err != nil {
			evalCtx.Log.Errorf(err.Error())
		}
		evalCtx.Log.Info("Goal results transacted")
	} else {
		evalCtx.Log.Info("No goal results to transact")
	}

	return resultEntity, nil
}
