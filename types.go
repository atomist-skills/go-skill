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

type EventContextSubscription struct {
	Name          string         `edn:"name"`
	Configuration Configuration  `edn:"configuration"`
	Result        edn.RawMessage `edn:"result"`
	Metadata      struct {
		AfterBasisT  int64  `edn:"after-basis-t"`
		Tx           int64  `edn:"tx"`
		ScheduleName string `edn:"schedule-name"`
	} `edn:"metadata"`
}

type EventContextWebhook struct {
	Name          string        `edn:"name"`
	Configuration Configuration `edn:"configuration"`
	Request       struct {
		Url     string            `edn:"url"`
		Body    string            `edn:"body"`
		Headers map[string]string `edn:"headers"`
		Tags    []ParameterValue  `edn:"tags"`
	} `edn:"request"`
}

type EventContextSyncRequest struct {
	Name          string         `edn:"name"`
	Configuration Configuration  `edn:"configuration"`
	Metadata      edn.RawMessage `edn:"metadata"`
}

type EventContextAsyncQueryResult struct {
	Name          string         `edn:"name"`
	Configuration Configuration  `edn:"configuration"`
	Metadata      string         `edn:"metadata"`
	Result        edn.RawMessage `edn:"result"`
}

type EventContextEvent struct {
	Name     string                         `edn:"name"`
	Metadata map[edn.Keyword]edn.RawMessage `edn:"metadata"`
}

type EventContext struct {
	Subscription     EventContextSubscription     `edn:"subscription"`
	Webhook          EventContextWebhook          `edn:"webhook"`
	SyncRequest      EventContextSyncRequest      `edn:"sync-request"`
	AsyncQueryResult EventContextAsyncQueryResult `edn:"query-result"`
	Event            EventContextEvent            `edn:"event"`
}

type EventIncoming struct {
	ExecutionId  string       `edn:"execution-id"`
	Skill        Skill        `edn:"skill"`
	Type         edn.Keyword  `edn:"type"`
	WorkspaceId  string       `edn:"workspace-id"`
	Environment  string       `edn:"environment,omitempty"`
	Organization string       `edn:"organization,omitempty"`
	Context      EventContext `edn:"context"`
	Urls         struct {
		Execution    string `edn:"execution"`
		Logs         string `edn:"logs"`
		Transactions string `edn:"transactions"`
		Query        string `edn:"query"`
		Graphql      string `edn:"graphql"`
		Entitlements string `edn:"entitlements"`
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
	State       edn.Keyword `edn:"state"`
	Reason      string      `edn:"reason,omitempty"`
	SyncRequest interface{} `edn:"sync-request,omitempty"`
}

type RequestContext struct {
	Event EventIncoming
	Log   Logger

	ctx context.Context
}

func (r *RequestContext) NewTransaction() Transaction {
	return NewTransactionFromRequest(r.ctx, r.Event, r.Log)
}

func NewTransactionFromRequest(ctx context.Context, event EventIncoming, logger Logger) Transaction {
	var sender messageSender
	if event.Type != "" {
		sender = createMessageSender(ctx, event, logger)
	} else {
		sender = createHttpMessageSender(event.WorkspaceId, event.Token, event.ExecutionId, logger)
	}

	transactor := func(entities []interface{}, ordered bool) error {
		if ordered {
			return sender.TransactOrdered(entities, event.ExecutionId)
		}

		return sender.Transact(entities)
	}

	return newTransaction(ctx, transactor)
}

type EventHandler func(ctx context.Context, req RequestContext) Status

type Handlers func(name string) (EventHandler, bool)

func HandlersFromMap(handlers map[string]EventHandler) Handlers {
	return func(name string) (EventHandler, bool) {
		handle, ok := handlers[name]
		return handle, ok
	}
}
