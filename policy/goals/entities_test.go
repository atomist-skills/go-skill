/*
 * Copyright © 2023 Atomist, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package goals

import (
	"testing"
	"time"

	"olympos.io/encoding/edn"
)

func TestCreateEntitiesFromResult(t *testing.T) {
	result := `[{:name "CVE-2023-2650", :details {:purl "pkg:alpine/openssl@3.1.0-r4?os_name=alpine&os_version=3.18", :cve "CVE-2023-2650", :severity "HIGH", :fixed-by "3.1.1-r0"} }]`

	resultModel := []GoalEvaluationQueryResult{}

	edn.Unmarshal([]byte(result), &resultModel)

	evaluationTs := time.Date(2023, 7, 10, 20, 1, 41, 0, time.UTC)

	entity := CreateEntitiesFromResults(resultModel, "test-definition", "test-configuration", "test-image", "storage-id", "config-hash", evaluationTs, 123)

	if entity.Definition != "test-definition" || entity.Configuration != "test-configuration" || entity.StorageId != "storage-id" || entity.CreatedAt.Format("2006-01-02T15:04:05.000Z") != "2023-07-10T20:01:41.000Z" {
		t.Errorf("metadata not set correctly")
	}

	if entity.DeviationCount != 1 {
		t.Errorf("incorrect number of deviations, expected %d, got %d", 1, entity.DeviationCount)
	}
}

func TestNoDataSetsRetraction(t *testing.T) {
	result := `[{:name "CVE-2023-2650", :details {:purl "pkg:alpine/openssl@3.1.0-r4?os_name=alpine&os_version=3.18", :cve "CVE-2023-2650", :severity "HIGH", :fixed-by "3.1.1-r0"} }]`

	resultModel := []GoalEvaluationQueryResult{}

	edn.Unmarshal([]byte(result), &resultModel)

	evaluationTs := time.Date(2023, 7, 10, 20, 1, 41, 0, time.UTC)

	entity := CreateEntitiesFromResults(resultModel, "test-definition", "test-configuration", "test-image", "no-data", "config-hash", evaluationTs, 123)

	if !entity.StorageId.(RetractionEntity).Retract || !entity.DeviationCount.(RetractionEntity).Retract {
		t.Errorf("metadata not set correctly")
	}
}
