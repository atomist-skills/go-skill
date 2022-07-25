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

type configuration struct {
	Name       string `edn:"name"`
	parameters []struct {
		Name  string      `edn:"name"`
		Value interface{} `edn:"value"`
	} `edn:"parameters"`
}

type EventIncoming struct {
	ExecutionId string `edn:"execution-id"`
	Skill       struct {
		Id        string `edn:"id"`
		Namespace string `edn:"namespace"`
		Name      string `edn:"name"`
		Version   string `edn:"version"`
	} `edn:"skill"`
	Type        edn.Keyword `edn:"type"`
	WorkspaceId string      `edn:"workspace-id"`
	Context     struct {
		Subscription struct {
			Name          string                             `edn:"name"`
			Configuration configuration                      `edn:"configuration"`
			AfterBasisT   int64                              `edn:"tx"`
			Tx            int64                              `edn:"after-basis-t"`
			Result        [][]map[edn.Keyword]edn.RawMessage `edn:"result"`
		} `edn:"subscription"`
		Webhook struct {
			Name          string        `edn:"name"`
			Configuration configuration `edn:"configuration"`
			Request       struct {
				Url     string            `edn:"url"`
				Body    string            `edn:"body"`
				Headers map[string]string `edn:"headers"`
				Tags    []struct {
					Name  string      `edn:"name"`
					Value interface{} `edn:"value"`
				} `edn:"tags"`
			} `edn:"request"`
		} `edn:"webhook"`
	} `edn:"context"`
	Urls struct {
		Execution    string `edn:"execution"`
		Logs         string `edn:"logs"`
		Transactions string `edn:"transactions"`
		Query        string `edn:"query"`
	} `edn:"urls"`
	Token string `edn:"token"`
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
