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

type SubscriptionIncoming struct {
	Name          string                             `edn:"name"`
	Configuration ConfigurationIncoming              `edn:"configuration"`
	AfterBasisT   int64                              `edn:"tx"`
	Tx            int64                              `edn:"after-basis-t"`
	Result        [][]map[edn.Keyword]edn.RawMessage `edn:"result"`
}

type Context struct {
	Subscription SubscriptionIncoming `edn:"subscription"`
}

type Urls struct {
	Execution    string `edn:"execution"`
	Logs         string `edn:"logs"`
	Transactions string `edn:"transactions"`
	Query        string `edn:"query"`
}

type EventIncoming struct {
	ExecutionId string      `edn:"execution-id"`
	Skill       Skill       `edn:"skill"`
	Type        edn.Keyword `edn:"type"`
	WorkspaceId string      `edn:"workspace-id"`
	Context     Context     `edn:"context"`
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
	Reason string      `edn:"reason,omitempty"`
}

type RequestContext struct {
	Event           EventIncoming
	Log             Logger
	Transact        Transact
	TransactOrdered TransactOrdered
}

type EventHandler func(ctx context.Context, req RequestContext) Status

type Handlers map[string]EventHandler
