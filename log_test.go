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
	"net/http"
	"net/http/httptest"
	"olympos.io/encoding/edn"
	"testing"
)

func TestSuccessfulLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("authorization")
		if authHeader != "Bearer token" {
			t.Errorf("Authorization header is wrong: %s", authHeader)
		}
		contentTypeHeader := req.Header.Get("content-type")
		if contentTypeHeader != "application/edn" {
			t.Errorf("Content-Type header is wrong: %s", contentTypeHeader)
		}
		var logEvent LogBody
		err := edn.NewDecoder(req.Body).Decode(&logEvent)
		if err != nil {
			t.Error(err)
		}
		if len(logEvent.Logs) != 1 {
			t.Errorf("Invalid number of log entries: %d", len(logEvent.Logs))
		}
		if logEvent.Logs[0].Text != "This is a test message" {
			t.Errorf("Invalid message: %s", logEvent.Logs[0].Text)
		}
		if logEvent.Logs[0].Level != Info {
			t.Errorf("Invalid level: %s", logEvent.Logs[0].Level)
		}
		rw.WriteHeader(202)
	}))
	defer server.Close()

	logger := createLogger(context.Background(), EventIncoming{
		Urls: struct {
			Execution    string `edn:"execution"`
			Logs         string `edn:"logs"`
			Transactions string `edn:"transactions"`
			Query        string `edn:"query"`
		}{
			Logs: server.URL,
		},
		Token: "token",
	})
	logger.Infof("This is a %s message", "test")
}
