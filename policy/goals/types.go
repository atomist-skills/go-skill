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
	"time"

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
		Details map[edn.Keyword]interface{} `edn:"details"`
	}

	DockerImageEntity struct {
		skill.Entity `entity-type:"docker/image"`
		Digest       string `edn:"docker.image/digest"`
	}

	RetractionEntity struct {
		Retract bool `edn:"retract"`
	}

	GoalEvaluationResultEntity struct {
		skill.Entity         `entity-type:"goal/result"`
		Definition           string                     `edn:"goal.definition/name"`
		Configuration        string                     `edn:"goal.configuration/name"`
		Subject              DockerImageEntity          `edn:"goal.result/subject"`
		DeviationCount       interface{}                `edn:"goal.result/deviation-count,omitempty"`
		StorageId            interface{}                `edn:"goal.result/storage-id,omitempty"`
		ConfigHash           string                     `edn:"goal.result/config-hash"`
		CreatedAt            time.Time                  `edn:"goal.result/created-at"`
		TransactionCondition TransactionConditionEntity `edn:"atomist/tx-iff"`
	}

	TransactionConditionEntity struct {
		Args  map[string]interface{} `edn:"args"`
		Where edn.RawMessage         `edn:"where"`
	}

	OsDistro struct {
		Name    string `edn:"os.distro/name"`
		Version string `edn:"os.distro/version"`
	}

	SubscriptionImage struct {
		Digest string    `edn:"docker.image/digest"`
		Distro *OsDistro `edn:"docker.image/distro"`
	}

	SubscriptionRepository struct {
		Host       string `edn:"docker.repository/host"`
		Repository string `edn:"docker.repository/repository"`
	}

	ImagePlatform struct {
		Architecture string `edn:"docker.platform/architecture" json:"architecture"`
		Os           string `edn:"docker.platform/os" json:"os"`
	}

	Attestation struct {
		PredicateType *string     `edn:"intoto.attestation/predicate-type"`
		Predicates    []Predicate `edn:"intoto.predicate/_attestation"`
	}

	Predicate struct {
		StartLine *int `edn:"slsa.provenance.from/start-line"` // if field is present then provenance is max-mode
	}

	ImageSubscriptionQueryResult struct {
		ImageDigest    string                  `edn:"docker.image/digest"`
		ImagePlatforms []ImagePlatform         `edn:"docker.image/platform" json:"platforms"`
		FromReference  *SubscriptionImage      `edn:"docker.image/from"`
		FromRepo       *SubscriptionRepository `edn:"docker.image/from-repository"`
		FromTag        string                  `edn:"docker.image/from-tag"`
		Attestations   []Attestation           `edn:"intoto.attestation/_subject"`
		User           string                  `edn:"docker.image/user,omitempty"`
	}

	EvaluationMetadata struct {
		SubscriptionResult []map[edn.Keyword]edn.RawMessage `edn:"subscription-result"`
		SubscriptionTx     int64                            `edn:"subscription-tx"`
	}

	GoalEvaluator interface {
		EvaluateGoal(ctx context.Context, req skill.RequestContext, sbom types.SBOM) (EvaluationResult, error)
	}

	EvaluationResult struct {
		EvaluationCompleted bool
		Result              []GoalEvaluationQueryResult
	}
)
