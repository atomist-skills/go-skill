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

type Configuration struct {
	Name       string           `edn:"name,omitempty"`
	Parameters []ParameterValue `edn:"parameters,omitempty"`
}

type Skill struct {
	Id        string `edn:"id,omitempty"`
	Namespace string `edn:"namespace,omitempty"`
	Name      string `edn:"name,omitempty"`
	Version   string `edn:"version,omitempty"`
}

type EventIncoming struct {
	ExecutionId string      `edn:"execution-id,omitempty"`
	Skill       Skill       `edn:"skill,omitempty"`
	Type        edn.Keyword `edn:"type,omitempty"`
	WorkspaceId string      `edn:"workspace-id,omitempty"`
	Context     struct {
		Subscription struct {
			Name          string                             `edn:"name,omitempty"`
			Configuration Configuration                      `edn:"configuration,omitempty"`
			Result        [][]map[edn.Keyword]edn.RawMessage `edn:"result,omitempty"`
			Metadata      struct {
				AfterBasisT  int64  `edn:"tx,omitempty"`
				Tx           int64  `edn:"after-basis-t,omitempty"`
				ScheduleName string `edn:"schedule-name,omitempty"`
			} `edn:"metadata,omitempty"`
		} `edn:"subscription,omitempty"`
		Webhook struct {
			Name          string        `edn:"name,omitempty"`
			Configuration Configuration `edn:"configuration,omitempty"`
			Request       struct {
				Url     string            `edn:"url,omitempty"`
				Body    string            `edn:"body,omitempty"`
				Headers map[string]string `edn:"headers,omitempty"`
				Tags    []ParameterValue  `edn:"tags,omitempty"`
			} `edn:"request,omitempty"`
		} `edn:"webhook,omitempty"`
	} `edn:"context,omitempty"`
	Urls struct {
		Execution    string `edn:"execution,omitempty"`
		Logs         string `edn:"logs,omitempty"`
		Transactions string `edn:"transactions,omitempty"`
		Query        string `edn:"query,omitempty"`
	} `edn:"urls,omitempty"`
	Token string `edn:"token,omitempty"`
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
