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
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/atomist-skills/go-skill/internal"

	"github.com/google/uuid"

	"olympos.io/encoding/edn"
)

// Entity models the required fields to transact an entity
type Entity struct {
	EntityType edn.Keyword `edn:"schema/entity-type"`
	Entity     string      `edn:"schema/entity,omitempty"`
}

// ManyRef models an entity reference of cardinality many
type ManyRef struct {
	Add     []string `edn:"add,omitempty"`
	Set     []string `edn:"set,omitempty"`
	Retract []string `edn:"retract,omitempty"`
}

// Transaction collects entities
type Transaction interface {
	Ordered() Transaction
	AddEntities(entities ...interface{}) Transaction
	EntityRefs(entityType string) []string
	EntityRef(entityType string) string
	Transact() error
}

type transaction struct {
	entities   []interface{}
	ctx        context.Context
	context    RequestContext
	ordered    bool
	transactor Transactor
}

// Ordered makes this ordered
func (t *transaction) Ordered() Transaction {
	t.ordered = true
	return t
}

// AddEntities adds a new entity to this transaction
func (t *transaction) AddEntities(entities ...interface{}) Transaction {
	for _, e := range entities {
		t.entities = append(t.entities, makeEntity(e))
	}
	return t
}

// Transact triggers a transaction of the entities to the backend.
// The recorded entities are not discarded in this transaction for further reference
func (t *transaction) Transact() error {
	if t.transactor != nil {
		transactions, err := makeTransaction(t.entities, "")
		if err != nil {
			return err
		}

		flattenedEntities := transactions.Data
		bs, err := edn.MarshalPPrint(flattenedEntities, nil)
		t.transactor(string(bs))
		return nil
	} else {
		var transactor messageSender
		if t.context.Event.Type != "" {
			transactor = createMessageSender(t.ctx, t.context)
		} else {
			transactor = createHttpMessageSender(t.context.Event.WorkspaceId, t.context.Event.Token)
		}
		if t.ordered {
			return transactor.TransactOrdered(t.entities, t.context.Event.ExecutionId)
		} else {
			return transactor.Transact(t.entities)
		}
	}
}

// MakeEntity creates a new Entity struct populated with entity-type and a unique entity identifier
func MakeEntity[E interface{}](value E, entityId ...string) E {
	reflectValue := reflect.ValueOf(value)
	var field reflect.StructField
	if reflectValue.Kind() == reflect.Ptr {
		field, _ = reflect.TypeOf(value).Elem().FieldByName("Entity")
	} else {
		field, _ = reflect.TypeOf(value).FieldByName("Entity")
	}
	entityType := field.Tag.Get("entity-type")

	entity := Entity{
		EntityType: edn.Keyword(entityType),
	}
	if len(entityId) == 0 {
		parts := strings.Split(entityType, "/")
		entity.Entity = fmt.Sprintf("$%s-%s", parts[len(parts)-1], uuid.New().String())
	} else {
		entity.Entity = entityId[0]
	}

	if reflectValue.Kind() == reflect.Ptr {
		reflectValue.Elem().FieldByName("Entity").Set(reflect.ValueOf(entity))
	} else {
		reflect.ValueOf(&value).Elem().FieldByName("Entity").Set(reflect.ValueOf(entity))
	}

	return value
}

func (t *transaction) EntityRefs(entityType string) []string {
	return EntityRefs(t.entities, entityType)
}

func (t *transaction) EntityRef(entityType string) string {
	return EntityRef(t.entities, entityType)
}

// newTransaction creates a new Transaction to record entities
func newTransaction(ctx context.Context, context RequestContext, transactor Transactor) Transaction {
	return &transaction{
		entities:   make([]interface{}, 0),
		ctx:        ctx,
		context:    context,
		ordered:    false,
		transactor: transactor,
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

type messageSender struct {
	Transact        Transact
	TransactOrdered TransactOrdered
}

func createMessageSender(ctx context.Context, req RequestContext) messageSender {
	messageSender := messageSender{}

	messageSender.TransactOrdered = func(entities interface{}, orderingKey string) error {
		// Don't transact when evaluating policies locally
		if os.Getenv("SCOUT_LOCAL_POLICY_EVALUATION") == "true" {
			return nil
		}

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

		bs, err := edn.MarshalPPrint(internal.TransactionEntityBody{
			Transactions: []internal.TransactionEntity{*transactions}}, nil)
		if err != nil {
			return err
		}

		client := &http.Client{}

		req.Log.Debugf("Transacting entities: %s", string(bs))

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.Event.Urls.Transactions, bytes.NewBuffer(bs))
		if err != nil {
			return err
		}
		httpReq.Header.Set("Authorization", "Bearer "+req.Event.Token)
		httpReq.Header.Set("Content-Type", "application/edn")
		resp, err := client.Do(httpReq)
		if err != nil {
			return err
		}
		if resp.StatusCode != 202 {
			Log.Warnf("Error transacting entities: %s", resp.Status)
		}
		defer resp.Body.Close()

		return nil
	}

	messageSender.Transact = func(entities interface{}) error {
		return messageSender.TransactOrdered(entities, "")
	}

	return messageSender
}

func makeTransaction(entities []interface{}, orderingKey string) (*internal.TransactionEntity, error) {
	body, err := edn.MarshalPPrint(entities, nil)
	if err != nil {
		return nil, err
	}

	var e []map[edn.Keyword]edn.RawMessage
	err = edn.NewDecoder(bytes.NewReader(body)).Decode(&e)
	if err != nil {
		return nil, err
	}

	transactions := internal.TransactionEntity{Data: flattenEntities(e)}
	if orderingKey != "" {
		transactions.OrderingKey = orderingKey
	}
	return &transactions, nil
}

func flattenEntities(entities []map[edn.Keyword]edn.RawMessage) []map[edn.Keyword]edn.RawMessage {
	fEntities := make([]map[edn.Keyword]edn.RawMessage, 0)
	for _, e := range entities {
		fEntities = append(fEntities, flattenEntity(e)...)
	}

	// make entity list unique by schema/entity
	uEntities := make(map[string]map[edn.Keyword]edn.RawMessage, 0)
	for _, e := range fEntities {
		entity := string(e["schema/entity"])
		if _, ok := uEntities[entity]; !ok {
			uEntities[entity] = e
		}
	}

	// collect the values
	fEntities = make([]map[edn.Keyword]edn.RawMessage, 0)
	for _, v := range uEntities {
		fEntities = append(fEntities, v)
	}

	return fEntities
}

func flattenEntity(entity map[edn.Keyword]edn.RawMessage) []map[edn.Keyword]edn.RawMessage {
	entities := make([]map[edn.Keyword]edn.RawMessage, 0)
	entities = append(entities, entity)

	for k, v := range entity {
		if !strings.HasPrefix(k.String(), ":schema/") {
			// test single entity first
			var n map[edn.Keyword]edn.RawMessage
			err := edn.NewDecoder(bytes.NewReader(v)).Decode(&n)
			if err == nil {
				if e, ok := n["schema/entity"]; ok {
					entity[k] = e
					entities = append(entities, flattenEntity(n)...)
				}
				continue
			}
			// test array second
			var a []map[edn.Keyword]edn.RawMessage
			err = edn.NewDecoder(bytes.NewReader(v)).Decode(&a)
			if err == nil {
				refs := make([]string, len(a))
				for i := range a {
					refs[i] = string(a[i]["schema/entity"])
					entities = append(entities, flattenEntity(a[i])...)
				}
				entity[k] = []byte(fmt.Sprintf("{:set [%s]}", strings.Join(refs, "")))
			}
		}
	}

	return entities
}

func makeEntity(x interface{}) interface{} {
	// Starting value must be a pointer.
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		v = reflect.ValueOf(&x)
	}
	setEntityValues(v, "")
	return x
}

func setEntityValues(v reflect.Value, entityType string) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsZero() {
			return
		}
		setEntityValues(v.Elem(), entityType)
	case reflect.Interface:
		if v.IsZero() {
			return
		}
		iv := v.Elem()
		switch iv.Kind() {
		case reflect.Slice, reflect.Ptr:
			setEntityValues(iv, entityType)
		case reflect.Struct, reflect.Array:
			// Copy required for modification.
			copy := reflect.New(iv.Type()).Elem()
			copy.Set(iv)
			setEntityValues(copy, entityType)
			v.Set(copy)
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			fv := v.Field(i)
			if sf.Name == "Entity" {
				if entityType != "" {
					if fv.String() == "" {
						parts := strings.Split(entityType, "/")
						fv.Set(reflect.ValueOf(fmt.Sprintf("$%s-%s", parts[len(parts)-1], uuid.New().String())))
					}
				} else {
					entityType := sf.Tag.Get("entity-type")
					setEntityValues(fv, entityType)
				}
			} else if sf.Name == "EntityType" {
				if fv.Interface().(edn.Keyword) == "" {
					fv.Set(reflect.ValueOf(edn.Keyword(entityType)))
				}
			} else {
				setEntityValues(fv, entityType)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			setEntityValues(v.Index(i), entityType)
		}

	}
}
