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

package test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

type SimulateOptions struct {
	Skill         skill.Skill
	Subscription  string
	Schemata      string
	Configuration skill.Configuration
	TxData        string
	WorkspaceId   string
	Token         string
}

type SimulateResult struct {
	Results []struct {
		ConfigurationName string             `edn:"configuration-name"`
		Subscription      string             `edn:"subscription"`
		Results           [][]edn.RawMessage `edn:"results"`
	} `edn:"results"`
}

func Simulate(options SimulateOptions, t *testing.T) SimulateResult {
	if options.WorkspaceId == "" {
		options.WorkspaceId = os.Getenv("ATOMIST_WORKSPACE")
	}
	if options.Token == "" {
		options.Token = os.Getenv("ATOMIST_API_KEY")
	}
	if options.Token == "" || options.WorkspaceId == "" {
		t.Logf("Missing workspace and/or API key. Either pass both via options or set $ATOMIST_WORKSPACE and $ATOMIST_API_KEY. Skipping test...")
		t.Skip()
	}

	subscription, err := os.ReadFile(options.Subscription)
	subscriptionName := filepath.Base(options.Subscription)
	if err != nil {
		t.Fatalf("Failed to load subscription: %s", err)
	}
	txData, err := os.ReadFile(options.TxData)
	if err != nil {
		t.Fatalf("Failed to load tx data: %s", err)
	}
	schemata := ""
	if options.Schemata != "" {
		schema, err := os.ReadFile(options.Schemata)
		schemaName := filepath.Base(options.Schemata)
		if err != nil {
			t.Fatalf("Failed to load schema: %s", err)
		}
		schemata = fmt.Sprintf(`:schemata [{:name "%s"
					:schema %s}]`, schemaName[0:len(schemaName)-4], string(schema))
	}

	configuration, err := edn.MarshalPPrint(options.Configuration, nil)

	payload := fmt.Sprintf(`{
 :skill {:id        "%s"
         :namespace "%s"
         :name      "%s"
         :version   "%s"
         %s
         :subscriptions [{:name "%s"
                          :query %s}]
         :configurations [%s]}
		
 :tx-data %s
}
`, options.Skill.Id, options.Skill.Namespace, options.Skill.Name, options.Skill.Version, schemata, subscriptionName[0:len(subscriptionName)-4], string(subscription), string(configuration), string(txData))

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.atomist.com/datalog/team/%s/simulate", options.WorkspaceId), strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+options.Token)
	req.Header.Set("Content-Type", "application/edn")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to simulate subscription: %s", err)
	}
	buf := new(strings.Builder)
	io.Copy(buf, resp.Body)
	fmt.Println(buf.String())

	var results SimulateResult
	err = edn.NewDecoder(strings.NewReader(buf.String())).Decode(&results)
	if err != nil {
		t.Fatalf("Failed to decode simulation results: %s", err)
	}

	return results
}
