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
	"github.com/google/uuid"
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

// ManyRef models a entity reference of cardinality many
type ManyRef struct {
	Add     []string `edn:"add,omitempty"`
	Set     []string `edn:"set,omitempty"`
	Retract []string `edn:"retract,omitempty"`
}

// Transaction collects entities
type Transaction interface {
	MakeEntity(entityType edn.Keyword, entityId ...string) Entity
	AddEntities(entities ...interface{})
	EntityRefs(entityType string) []string
	EntityRef(entityType string) string
	Entities() []interface{}
}

type transaction struct {
	entities []interface{}
}

// AddEntities adds a new entity to this transaction
func (t *transaction) AddEntities(entities ...interface{}) {
	t.entities = append(entities, t.entities...)
}

// Entities returns all current entities in this transaction
func (t *transaction) Entities() []interface{} {
	return t.entities
}

// MakeEntity creates a new Entity struct populated with all values
func (t *transaction) MakeEntity(entityType edn.Keyword, entityId ...string) Entity {
	entity := Entity{
		EntityType: entityType,
	}
	if len(entityId) == 0 {
		entity.Entity = "$" + uuid.New().String()
	} else {
		entity.Entity = entityId[0]
	}
	return entity
}

func (t *transaction) EntityRefs(entityType string) []string {
	return EntityRefs(t.entities, entityType)
}

func (t *transaction) EntityRef(entityType string) string {
	return EntityRef(t.entities, entityType)
}

// NewTransaction creates a new Transaction to record entities
func NewTransaction() Transaction {
	return &transaction{
		entities: make([]interface{}, 0),
	}
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

type TransactionEntity struct {
	Data        []interface{} `edn:"data"`
	OrderingKey string        `edn:"ordering-key,omitempty"`
}

type TransactEntityBody struct {
	Transactions []TransactionEntity `edn:"transactions"`
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

		var transactions = TransactionEntity{Data: entityArray}
		if orderingKey != "" {
			transactions.OrderingKey = orderingKey
		}

		bs, err := edn.MarshalPPrint(TransactEntityBody{
			Transactions: []TransactionEntity{transactions}}, nil)

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
