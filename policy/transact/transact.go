package transact

import (
	"context"
	"fmt"
	"time"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/storage"
	"github.com/atomist-skills/go-skill/util"
	"olympos.io/encoding/edn"
)

func TransactPolicyResult(
	ctx context.Context,
	evalCtx goals.GoalEvaluationContext,
	configuration skill.Configuration,
	goalName string,
	digest string,
	goal goals.Goal,
	subscriptionResult []map[edn.Keyword]edn.RawMessage,
	evaluationTs time.Time,
	goalResults []goals.GoalEvaluationQueryResult,
	tx int64,
) skill.Status {
	storageTuple := util.Decode[[]string](subscriptionResult[0]["previous"])
	previousStorageId := storageTuple[0]
	previousConfigHash := storageTuple[1]

	if goalResults == nil {
		req.Log.Infof("goal %s returned no data for digest %s", goal.Definition, digest)
	}

	es, err := storage.NewEvaluationStorage(ctx)
	if err != nil {
		return skill.NewFailedStatus(fmt.Sprintf("Failed to create evaluation storage: %s", err.Error()))
	}

	configDiffer, configHash, err := goals.GoalConfigsDiffer(req.Log, configuration, digest, goal, previousConfigHash)
	if err != nil {
		req.Log.Errorf("Failed to check if config hash changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		configDiffer = true
	}

	differ, storageId, err := goals.GoalResultsDiffer(req.Log, goalResults, digest, goal, previousStorageId)
	if err != nil {
		req.Log.Errorf("Failed to check if goal results changed for digest: %s", digest, err)
		req.Log.Warnf("Will continue with the evaluation nonetheless")
		differ = true
	}

	if differ && goalResults != nil {
		if err := es.Store(ctx, goalResults, storageId, req.Event.Environment, req.Log); err != nil {
			return skill.NewFailedStatus(fmt.Sprintf("Failed to store evaluation results for digest %s: %s", digest, err.Error()))
		}
	}

	var entities []interface{}
	if differ || configDiffer {
		shouldRetract := previousStorageId != "no-data" && previousStorageId != "n/a" && storageId == "no-data"
		entity := goals.CreateEntitiesFromResults(goalResults, goal.Definition, goal.Configuration, digest, storageId, configHash, evaluationTs, tx, shouldRetract)
		entities = append(entities, entity)
	}

	if len(entities) > 0 {
		err = req.NewTransaction().AddEntities(entities...).Transact()
		if err != nil {
			req.Log.Errorf(err.Error())
		}
		req.Log.Info("Goal results transacted")
	} else {
		req.Log.Info("No goal results to transact")
	}

	return skill.NewCompletedStatus(fmt.Sprintf("Goal %s evaluated", goalName))
}