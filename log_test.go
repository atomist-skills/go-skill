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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atomist-skills/go-skill/internal"
	"github.com/sirupsen/logrus"

	"olympos.io/encoding/edn"
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
		var logEvent internal.LogBody
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
		if logEvent.Logs[0].Level != internal.Info {
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
			Graphql      string `edn:"graphql"`
			Entitlements string `edn:"entitlements"`
		}{
			Logs: server.URL,
		},
		Token: "token",
	}, http.Header{}, nil)
	logger.Infof("This is a %s message", "test")
}

func TestSanitizeEvent(t *testing.T) {
	var payload = "{:execution-id \"855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg\", :skill {:namespace \"atomist\", :name \"go-sample-skill\", :version \"0.1.0-100\"}, :workspace-id \"T29E48P34\", :type :subscription, :context {:subscription {:name \"on_push\", :configuration {:name \"go_sample_skill\", :parameters [{:name \"repoFilter\", :value {}} {:name \"on_webhook\", :value ({:name \"on_webhook-0\", :url \"https://webhook.atomist.com/atomist/resource/b36b6db3-7d73-442b-9809-626a9ce036d0\"})}]}, :result ([{:git.commit/repo {:git.repo/name \"go-sample-skill\", :git.repo/source-id \"490643782\", :git.repo/default-branch \"main\", :git.repo/org {:github.org/installation-token \"ghs_H9bCqKtdsdfsfsdfsfsfsfQ8BeD6iWrSGM4RfYZm\", :git.org/name \"atomist-skills\", :git.provider/url \"https://github.com\"}}, :git.commit/author {:git.user/name \"atomist[bot]\", :git.user/login \"atomist[bot]\", :git.user/emails [{:email.email/address \"22779605+atomist[bot]@users.noreply.github.com\"}]}, :git.commit/sha \"8969fcce08a2869affc001a05fd8471bcf92b28f\", :git.commit/message \"Auto-merge pull request #21 from atomist-skills/go-sample-skill\", :git.ref/refs [{:git.ref/name \"main\", :git.ref/type {:db/id 83562883711320, :db/ident :git.ref.type/branch}}]}]), :metadata {:after-basis-t 4354969, :tx 13194143888281}, :after-basis-t 4354969, :tx 13194143888281}}, :urls {:execution \"https://api.atomist.com/executions/855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg\", :logs \"https://api.atomist.com/executions/855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg/logs\", :transactions \"https://api.atomist.com/executions/855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg/transactions\", :query \"https://api.atomist.com/datalog/team/T29E48P34/queries\"}, :token \"eyJhbGciOiJSUzI1NiOGd_6YHE8ud8GsBMy4E\"}"
	sanitizedEvent := sanitizeEvent(payload)
	if strings.Contains(sanitizedEvent, "ghs_H9bCqKtdsdfsfsdfsfsfsfQ8BeD6iWrSGM4RfYZm") {
		t.Errorf("installation token not sanitized")
	}
	if strings.Contains(sanitizedEvent, "eyJhbGciOiJSUzI1NiOGd_6YHE8ud8GsBMy4E") {
		t.Errorf("token not sanitized")
	}
}

func TestSanitizeEventWithSingleCharacterUser(t *testing.T) {
	var payload = "{:execution-id \"855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg\", :skill {:namespace \"atomist\", :name \"go-sample-skill\", :version \"0.1.0-100\"}, :workspace-id \"T29E48P34\", :type :subscription, :context {:subscription {:name \"on_push\", :configuration {:name \"go_sample_skill\", :parameters [{:name \"repoFilter\", :value {}} {:name \"on_webhook\", :value ({:name \"on_webhook-0\", :url \"https://webhook.atomist.com/atomist/resource/b36b6db3-7d73-442b-9809-626a9ce036d0\"})}]}, :result ([{:git.commit/repo {:git.repo/name \"go-sample-skill\", :git.repo/source-id \"490643782\", :git.repo/default-branch \"main\", :git.repo/org {:github.org/installation-token \"ghs_H9bCqKtdsdfsfsdfsfsfsfQ8BeD6iWrSGM4RfYZm\", :git.org/name \"atomist-skills\", :git.provider/url \"https://github.com\"}}, :git.commit/author {:git.user/name \"0\", :git.user/login \"atomist[bot]\", :git.user/emails [{:email.email/address \"22779605+atomist[bot]@users.noreply.github.com\"}]}, :git.commit/sha \"8969fcce08a2869affc001a05fd8471bcf92b28f\", :git.commit/message \"Auto-merge pull request #21 from atomist-skills/go-sample-skill\", :git.ref/refs [{:git.ref/name \"main\", :git.ref/type {:db/id 83562883711320, :db/ident :git.ref.type/branch}}]}]), :metadata {:after-basis-t 4354969, :tx 13194143888281}, :after-basis-t 4354969, :tx 13194143888281}}, :urls {:execution \"https://api.atomist.com/executions/855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg\", :logs \"https://api.atomist.com/executions/855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg/logs\", :transactions \"https://api.atomist.com/executions/855f5639-8627-4bf2-86e8-51346019ddcb.iStU3P05jAeiKAJ7pnXfg/transactions\", :query \"https://api.atomist.com/datalog/team/T29E48P34/queries\"}, :token \"eyJhbGciOiJSUzI1NiOGd_6YHE8ud8GsBMy4E\"}"
	sanitizedEvent := sanitizeEvent(payload)

	if strings.Contains(sanitizedEvent, "\"0\"") {
		t.Errorf("user not sanitised")
	}
}

func TestLoggingWithFunc(t *testing.T) {
	var buf bytes.Buffer
	Log.SetOutput(&buf)
	Log.SetLevel(logrus.DebugLevel)
	logger := createLogger(context.Background(), EventIncoming{}, http.Header{}, nil)
	logger.Debugf("This is a %s message", func() interface{} { return "test" })

	if !strings.Contains(buf.String(), "This is a test message") {
		t.Errorf("Expected message not found")
	}
}
