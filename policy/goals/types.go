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
	"context"
	"github.com/atomist-skills/go-skill/policy/query"
	"time"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

type Goal struct {
	Args          map[string]interface{}
	Definition    string
	Configuration string
}

type GoalEvaluationQueryResult struct {
	Details map[edn.Keyword]interface{} `edn:"details"`
}

type DockerImageEntity struct {
	skill.Entity `entity-type:"docker/image"`
	Digest       string `edn:"docker.image/digest"`
}

type GoalEvaluationResultEntity struct {
	skill.Entity   `entity-type:"goal/result"`
	Definition     string            `edn:"goal.definition/name"`
	Configuration  string            `edn:"goal.configuration/name"`
	Subject        DockerImageEntity `edn:"goal.result/subject"`
	DeviationCount int               `edn:"goal.result/deviation-count"`
	StorageId      string            `edn:"goal.result/storage-id"`
	ConfigHash     string            `edn:"goal.result/config-hash"`
	CreatedAt      time.Time         `edn:"goal.result/created-at"`
}

type GoalEvaluator interface {
	EvaluateGoal(ctx context.Context, req skill.RequestContext, commonData query.CommonSubscriptionQueryResult, subscriptionResults [][]edn.RawMessage) ([]GoalEvaluationQueryResult, error)
}
