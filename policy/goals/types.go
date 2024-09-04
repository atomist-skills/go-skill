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

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/types"
	"olympos.io/encoding/edn"
)

type (
	Goal struct {
		Args          map[string]interface{}
		Definition    string
		Configuration string
	}

	GoalEvaluationQueryResult struct {
		Details map[edn.Keyword]interface{} `edn:"details" json:"details"`
	}

	GoalEvaluator interface {
		EvaluateGoal(ctx context.Context, evalCtx GoalEvaluationContext, sbom types.SBOM, extraData []map[edn.Keyword]edn.RawMessage) (EvaluationResult, error)
	}

	EvaluationResult struct {
		EvaluationCompleted bool
		Result              []GoalEvaluationQueryResult
	}

	GoalEvaluationContext struct {
		Log          skill.Logger
		TeamId       string
		Organization string
		Goal         Goal
	}
)
