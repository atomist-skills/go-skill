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
	"log"
	"olympos.io/encoding/edn"
	"strings"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	var payload = "{:execution-id\n   \"698e4c21-bf56-482b-be70-54273910fc37.YgDUPm3oIpDTT0SbYH5t5\"\n :skill\n   {:namespace \"atomist\"\n    :name \"go-sample-skill\"\n    :version \"0.1.0-42\"}\n :workspace-id \"T29E48P34\"\n :type :subscription\n :context\n   {:subscription\n      {:name \"on_push\"\n       :configuration {:name \"go_sample_skill\"}\n       :result\n         ([{:schema/entity-type \":git/commit\"\n            :git.commit/repo\n              {:git.repo/name \"go-sample-skill\"\n               :git.repo/source-id \"490643782\"\n               :git.repo/default-branch \"main\"\n               :git.repo/org\n                 {:github.org/installation-token\n                    \"ghs_jrb2OZetoqnpa7v27NNj2egKzOPniB1U8YTH\"\n                  :git.org/name \"atomist-skills\"\n                  :git.provider/url \"https://github.com\"}}\n            :git.commit/author\n              {:git.user/name \"Christian Dupuis\"\n               :git.user/login \"cdupuis\"\n               :git.user/emails\n                 [{:email.email/address \"cd@atomist.com\"}]}\n            :git.commit/sha\n              \"68c3d821eddc46c4dc4b1de0ffb1a6c29a5342a9\"\n            :git.commit/message \"Update README.md\"\n            :git.ref/refs\n              [{:git.ref/name \"main\"\n                :git.ref/type\n                  {:db/id 83562883711320\n                   :db/ident \":git.ref.type/branch\"}}]}])\n       :after-basis-t 4284274\n       :tx 13194143817586}}\n :urls\n   {:execution\n      \"https://api.atomist.com/executions/698e4c21-bf56-482b-be70-54273910fc37.YgDUPm3oIpDTT0SbYH5t5\"\n    :logs\n      \"https://api.atomist.com/executions/698e4c21-bf56-482b-be70-54273910fc37.YgDUPm3oIpDTT0SbYH5t5/logs\"\n    :transactions\n      \"https://api.atomist.com/executions/698e4c21-bf56-482b-be70-54273910fc37.YgDUPm3oIpDTT0SbYH5t5/transactions\"\n    :query\n      \"https://api.atomist.com/datalog/team/T29E48P34/queries\"}\n :token \"[JSON_WEB_TOKEN]\"}"
	var event EventIncoming
	err := edn.NewDecoder(strings.NewReader(payload)).Decode(&event)
	name := nameFromEvent(event)
	if name != "on_push" {
		t.Errorf("Expected on_push as name")
	}
	if err != nil {
		log.Fatalln(err)
	}
}
