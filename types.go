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
	Name       string           `edn:"name"`
	Parameters []ParameterValue `edn:"parameters,omitempty"`
}

type Skill struct {
	Id        string `edn:"id"`
	Namespace string `edn:"namespace"`
	Name      string `edn:"name"`
	Version   string `edn:"version"`
}

type EventIncoming struct {
	ExecutionId string      `edn:"execution-id"`
	Skill       Skill       `edn:"skill"`
	Type        edn.Keyword `edn:"type"`
	WorkspaceId string      `edn:"workspace-id"`
	Context     struct {
		Subscription struct {
			Name          string                             `edn:"name"`
			Configuration Configuration                      `edn:"configuration"`
			Result        [][]map[edn.Keyword]edn.RawMessage `edn:"result"`
			Metadata      struct {
				AfterBasisT  int64  `edn:"tx"`
				Tx           int64  `edn:"after-basis-t"`
				ScheduleName string `edn:"schedule-name"`
			} `edn:"metadata"`
		} `edn:"subscription"`
		Webhook struct {
			Name          string        `edn:"name"`
			Configuration Configuration `edn:"configuration"`
			Request       struct {
				Url     string            `edn:"url"`
				Body    string            `edn:"body"`
				Headers map[string]string `edn:"headers"`
				Tags    []ParameterValue  `edn:"tags"`
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
	queued    edn.Keyword = "queued"
	running   edn.Keyword = "running"
	Completed edn.Keyword = "completed"
	retryable edn.Keyword = "retryable"
	Failed    edn.Keyword = "failed"
)

type Status struct {
	State  edn.Keyword `edn:"state"`
	Reason string      `edn:"reason,omitempty"`
}

type RequestContext struct {
	Event EventIncoming
	Log   Logger

	ctx context.Context
}

func (r *RequestContext) NewTransaction() Transaction {
	return newTransaction(r.ctx, *r, nil)
}

type Transactor func(entities string)

func (r *RequestContext) NewTransactionWithTransactor(transactor Transactor) Transaction {
	return newTransaction(r.ctx, *r, transactor)
}

type EventHandler func(ctx context.Context, req RequestContext) Status

type Handlers map[string]EventHandler
