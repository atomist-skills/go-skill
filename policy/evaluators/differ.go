package evaluators

import (
	"fmt"

	"github.com/atomist-skills/go-skill/policy/goals"

	"github.com/atomist-skills/go-skill"
	"github.com/mitchellh/hashstructure/v2"
)

// GoalResultsDiffer checks if the current query results differ from the previous ones.
// It returns the storage id for the current query results.
func GoalResultsDiffer(log skill.Logger, queryResults []goals.GoalEvaluationQueryResult, digest string, goal goals.Goal, previousStorageId string) (bool, string, error) {
	log.Infof("Generating storage id for goal %s, image %s", goal.Definition, digest)
	hash, err := hashstructure.Hash(queryResults, hashstructure.FormatV2, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to generate storage id for goal %s, image %s: %s", goal.Definition, digest, err)
	}

	storageId := fmt.Sprint(hash)

	differ := storageId != previousStorageId

	if differ {
		log.Infof("New storage id [%s] differs from previous [%s]", storageId, previousStorageId)
	} else {
		log.Infof("New storage id matches previous [%s]", storageId)
	}

	return differ, storageId, nil
}

func isRelevantParam(str string) bool {
	irrelevantParams := []string{"definitionName", "displayName", "description"}
	for _, v := range irrelevantParams {
		if v == str {
			return false
		}
	}

	return true
}

// Returns the config hash for the current skill config
func GoalConfigsDiffer(log skill.Logger, config skill.Configuration, digest string, goal goals.Goal, previousConfigHash string) (bool, string, error) {
	log.Debugf("Generating config hash for goal %s, image %s", goal.Definition, digest)

	params := config.Parameters
	values := map[string]interface{}{}
	for _, p := range params {
		if isRelevantParam(p.Name) {
			values[p.Name] = p.Value
		}
	}

	hash, err := hashstructure.Hash(values, hashstructure.FormatV2, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to generate config hash for goal %s, image %s: %s", goal.Definition, digest, err)
	}

	configHash := fmt.Sprint(hash)

	differ := configHash != previousConfigHash

	if differ {
		log.Infof("New config hash [%s] differs from previous [%s]", configHash, previousConfigHash)
	} else {
		log.Infof("New config hash matches previous [%s]", configHash)
	}

	return differ, configHash, nil
}
