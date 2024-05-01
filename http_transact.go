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
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/atomist-skills/go-skill/environment"
	"github.com/atomist-skills/go-skill/internal"
	"github.com/google/uuid"
	"olympos.io/encoding/edn"
)

func createHttpMessageSender(workspace string, apikey string, correlationId string, logger Logger) messageSender {
	return messageSender{
		Transact: func(entities interface{}) error {
			return httpTransact(entities, "", workspace, apikey, correlationId, logger)
		},
		TransactOrdered: func(entities interface{}, orderingKey string) error {
			return httpTransact(entities, orderingKey, workspace, apikey, correlationId, logger)
		},
	}
}

func httpTransact(entities interface{}, orderingKey string, workspace string, apikey string, correlationId string, logger Logger) error {
	var entityArray []interface{}
	rt := reflect.TypeOf(entities)
	switch rt.Kind() {
	case reflect.Array:
	case reflect.Slice:
		entityArray = entities.([]interface{})
	default:
		entityArray = []any{entities}
	}

	transactions, err := makeTransaction(entityArray, orderingKey)
	if err != nil {
		return err
	}

	flattenedEntities := transactions.Data
	bs, err := edn.MarshalPPrint(flattenedEntities, nil)
	if err != nil {
		return err
	}

	message := internal.ResponseMessage{
		ApiVersion:    "2",
		CorrelationId: "",
		Team: internal.Team{
			Id: workspace,
		},
		Type:     "facts_ingestion",
		Entities: string(bs),
	}

	if correlationId != "" {
		message.CorrelationId = correlationId
	} else if orderingKey != "" {
		message.CorrelationId = orderingKey
	} else {
		message.CorrelationId = uuid.NewString()
	}

	client := &http.Client{}

	logger.Debugf("Transacting entities with correlation id %s:\n%s", message.CorrelationId, string(bs))
	j, _ := json.MarshalIndent(message, "", "  ")

	url := "https://api.atomist.com/skills/remote/" + workspace
	if environment.IsStaging() {
		url = "https://api-staging.atomist.services/skills/remote/" + workspace
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apikey)
	httpReq.Header.Set("x-atomist-correlation-id", message.CorrelationId)
	if orderingKey != "" {
		httpReq.Header.Set("x-atomist-ordering-key", message.CorrelationId)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		logger.Warnf("Error transacting entities: %s", resp.Status)
	}

	return nil
}
