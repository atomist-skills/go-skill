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

package internal

import "olympos.io/encoding/edn"

const (
	Debug edn.Keyword = "debug"
	Info              = "info"
	Warn              = "warn"
	Error             = "error"
)

type LogEntry struct {
	Timestamp string      `edn:"timestamp"`
	Level     edn.Keyword `edn:"level"`
	Text      string      `edn:"text"`
}

type LogBody struct {
	Logs []LogEntry `edn:"logs"`
}

type TransactionEntity struct {
	Data        []map[edn.Keyword]edn.RawMessage `edn:"data"`
	OrderingKey string                           `edn:"ordering-key,omitempty"`
}

type TransactionEntityBody struct {
	Transactions []TransactionEntity `edn:"transactions"`
}

type StatusBody struct {
	Status interface{} `edn:"status,omitempty"`
}
