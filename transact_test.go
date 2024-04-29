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
	"testing"
)

type Foo struct {
	Entity `entity-type:"foo"`
	Bars   []Bar   `edn:"bars"`
	Bar    Bar     `edn:"bar"`
	Refs   ManyRef `edn:"refs"`
}

type Bar struct {
	Entity `entity-type:"bar"`
	Name   string `edn:"name"`
}

func TestMakeWithoutId(t *testing.T) {
	entity := MakeEntity(Foo{})
	if entity.Entity.Entity == "" {
		t.Errorf("Expected entity to be set")
	}
	if entity.Entity.Entity[0:1] != "$" {
		t.Errorf("Expected entity should start with $")
	}
}

func TestMakeWithId(t *testing.T) {
	entity := MakeEntity(Foo{}, "$repo")
	if entity.Entity.Entity != "$repo" {
		t.Errorf("Expected entity to be set to $repo")
	}
}

func TestEntityRefs(t *testing.T) {
	transaction := NewTransaction(context.TODO(), NewStringTransactor(func(entities string) {}))
	entity1 := Bar{Name: "1"}
	entity2 := Bar{Name: "2"}
	entity3 := Bar{Name: "3"}
	entity4 := Foo{}
	transaction.AddEntities(entity1, entity2).AddEntities(entity3, entity4)
	refs := transaction.EntityRefs("foo")
	if len(refs) != 1 {
		t.Errorf("Expected one entity ref")
	}
}

func TestMakeTransactionWithNested(t *testing.T) {
	foos := []any{Foo{
		Bars: []Bar{{
			Name: "Murphy's",
		}, {
			Name: "Irish Pub",
		}},
		Bar: Bar{
			Entity: Entity{
				EntityType: "barbar",
				Entity:     "test",
			},
			Name: "Kanapee",
		},
		Refs: ManyRef{Add: []string{"foo", "bar"}},
	}}

	newFoos := makeEntity(foos)

	transactionEntity, err := makeTransaction(newFoos.([]interface{}), "")
	if err != nil {
		t.Failed()
	}
	if len(transactionEntity.Data) != 4 {
		t.Errorf("Incorrect number of entities in transaction")
	}
}
