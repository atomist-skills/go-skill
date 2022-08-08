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
	"log"
	"net/http"
	"reflect"

	"olympos.io/encoding/edn"
)

// Entity models the required fields to transact an entity
type Entity struct {
	EntityType edn.Keyword `edn:"schema/entity-type"`
	Entity     string      `edn:"schema/entity,omitempty"`
}

// EntityRefs find all entities by given entityType and returns their identity
func EntityRefs(entities []interface{}, entityType string) []string {
	entityRefs := make([]string, 0)
	for i := range entities {
		entity := entities[i]
		if entity != nil && reflect.ValueOf(entity).FieldByName("EntityType").String() == entityType {
			value := reflect.ValueOf(entity).FieldByName("Entity").Interface().(Entity)
			entityRefs = append([]string{value.Entity}, entityRefs...)
		}
	}
	return entityRefs
}

// EntityRef finds one entity by given entityType and returns its identity
func EntityRef(entities []interface{}, entityType string) string {
	if entityRefs := EntityRefs(entities, entityType); len(entityRefs) > 0 {
		return entityRefs[0]
	}
	return ""
}

type Transact func(entities interface{}) error
type TransactOrdered func(entities interface{}, orderingKey string) error

type MessageSender struct {
	Transact        Transact
	TransactOrdered TransactOrdered
}

type Transaction struct {
	Data        []interface{} `edn:"data"`
	OrderingKey string        `edn:"ordering-key,omitempty"`
}

type TransactBody struct {
	Transactions []Transaction `edn:"transactions"`
}

func createMessageSender(ctx context.Context, req RequestContext) MessageSender {
	messageSender := MessageSender{}

	messageSender.TransactOrdered = func(entities interface{}, orderingKey string) error {
		var entityArray []interface{}
		rt := reflect.TypeOf(entities)
		switch rt.Kind() {
		case reflect.Array:
		case reflect.Slice:
			entityArray = entities.([]interface{})
		default:
			entityArray = []any{entities}
		}

		var transactions = Transaction{Data: entityArray}
		if orderingKey != "" {
			transactions.OrderingKey = orderingKey
		}

		bs, err := edn.MarshalPPrint(TransactBody{
			Transactions: []Transaction{transactions}}, nil)

		if err != nil {
			return err
		}

		client := &http.Client{}

		req.Log.Debugf("Transacting entities: %s", string(bs))

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.Event.Urls.Transactions, bytes.NewBuffer(bs))
		httpReq.Header.Set("Authorization", "Bearer "+req.Event.Token)
		httpReq.Header.Set("Content-Type", "application/edn")
		if err != nil {
			return err
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return err
		}
		if resp.StatusCode != 202 {
			log.Printf("Error transacting entities: %s", resp.Status)
		}
		defer resp.Body.Close()

		return nil
	}

	messageSender.Transact = func(entities interface{}) error {
		return messageSender.TransactOrdered(entities, "")
	}

	return messageSender
}
