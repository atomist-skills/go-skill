package goals

import (
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/mitchellh/hashstructure/v2"
)

// GoalResultsDiffer checks if the current query results differ from the previous ones.
// It returns the storage id for the current query results.
func GoalResultsDiffer(log skill.Logger, queryResults []GoalEvaluationQueryResult, digest string, previousStorageId string) (bool, string, error) {
	log.Infof("Generating storage id for image %s", digest)

	storageId := "no-data"

	if queryResults != nil {
		hashOptions := hashstructure.HashOptions{
			SlicesAsSets: true,
		}
		hash, err := hashstructure.Hash(queryResults, hashstructure.FormatV2, &hashOptions)
		if err != nil {
			return false, "", fmt.Errorf("failed to generate storage id for image %s: %s", digest, err)
		}

		storageId = fmt.Sprint(hash)
	}

	differ := storageId != previousStorageId

	if differ {
		log.Infof("New storage id [%s] differs from previous [%s]", storageId, previousStorageId)
	} else {
		log.Infof("New storage id matches previous [%s]", storageId)
	}

	return differ, storageId, nil
}

func isRelevantParam(str string) bool {
	irrelevantParams := []string{"definitionName", "displayName", "description", "remediationLink", "resultType", "detailsOrder"}
	for _, v := range irrelevantParams {
		if v == str {
			return false
		}
	}

	return true
}

// Returns the config hash for the current skill config
func GoalConfigsDiffer(log skill.Logger, config skill.Configuration, digest string, previousConfigHash string) (bool, string, error) {
	log.Debugf("Generating config hash for image %s", digest)

	params := config.Parameters
	values := map[string]interface{}{}
	for _, p := range params {
		if isRelevantParam(p.Name) {
			values[p.Name] = p.Value
		}
	}

	hashOptions := hashstructure.HashOptions{
		SlicesAsSets: true,
	}
	hash, err := hashstructure.Hash(values, hashstructure.FormatV2, &hashOptions)
	if err != nil {
		return false, "", fmt.Errorf("failed to generate config hash for image %s: %s", digest, err)
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
