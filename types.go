/*
 * Copyright © 2022 Atomist, Inc.
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

package skill

import (
	"context"
	"olympos.io/encoding/edn"
)

type ParameterValue struct {
	Name  string      `edn:"name"`
	Value interface{} `edn:"value"`
}

type ConfigurationIncoming struct {
	Name       string           `edn:"name"`
	parameters []ParameterValue `edn:"parameters"`
}

type SubscriptionIncoming[T any] struct {
	Name          string                `edn:"name"`
	Configuration ConfigurationIncoming `edn:"configuration"`
	AfterBasisT   int64                 `edn:"tx"`
	Tx            int64                 `edn:"after-basis-t"`
	Result        []T                   `edn:"result"`
}

type Context[T any] struct {
	Subscription SubscriptionIncoming[T] `edn:"subscription"`
}

type Urls struct {
	Execution   string `edn:"execution"`
	Logs        string `edn:"logs"`
	Transaction string `edn:"transaction"`
	Query       string `edn:"query"`
}

type EventIncoming[T any] struct {
	ExecutionId string      `edn:"execution-id"`
	Skill       Skill       `edn:"skill"`
	Type        edn.Keyword `edn:"type"`
	WorkspaceId string      `edn:"workspace-id"`
	Context     Context[T]  `edn:"context"`
	Urls        Urls        `edn:"urls"`
	Token       string      `edn:"token"`
}

type Skill struct {
	Id        string `edn:"id"`
	Namespace string `edn:"namespace"`
	Name      string `edn:"name"`
	Version   string `edn:"version"`
}

const (
	Queued    edn.Keyword = "queued"
	Running               = "running"
	Completed             = "completed"
	Retryable             = "retryable"
	Failed                = "failed"
)

type Status struct {
	State  edn.Keyword `edn:"state"`
	Reason string      `edn:"reason"`
}

type EventContext[T any] struct {
	Event    EventIncoming[T]
	Log      Logger
	Transact Transact
	Context  context.Context
}

type EventHandler[T any] func(ctx EventContext[T]) Status

type Handlers map[string]interface{}
