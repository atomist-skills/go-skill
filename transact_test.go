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

import "testing"

type testEntity struct {
	Entity
	Foo string
}

func TestMakeWithoutId(t *testing.T) {
	transaction := NewTransaction(RequestContext{}, false)
	entity := transaction.MakeEntity("foo")
	if entity.Entity == "" {
		t.Errorf("Expected entity to be set")
	}
	if entity.Entity[0:1] != "$" {
		t.Errorf("Expected entity should start with $")
	}
}

func TestMakeWithId(t *testing.T) {
	transaction := NewTransaction(RequestContext{}, false)
	entity := transaction.MakeEntity("foo", "$repo")
	if entity.Entity != "$repo" {
		t.Errorf("Expected entity to be set to $repo")
	}
}

func TestAddEntities(t *testing.T) {
	transaction := NewTransaction(RequestContext{}, false)
	entity1 := testEntity{
		Entity: transaction.MakeEntity("foo"),
	}
	entity2 := testEntity{
		Entity: transaction.MakeEntity("bar"),
	}
	entity3 := testEntity{
		Entity: transaction.MakeEntity("bar"),
	}
	transaction.AddEntities(entity1, entity2)
	transaction.AddEntities(entity3)
	if len(transaction.Entities()) != 3 {
		t.Errorf("Expected three entities")
	}
	refs := transaction.EntityRefs("foo")
	if len(refs) != 1 {
		t.Errorf("Expected one entity ref")
	}
}

type Foo struct {
	Entity
	Bars []Bar   `edn:"bars"`
	Refs ManyRef `edn:"refs""`
}

type Bar struct {
	Entity
	Name string `edn:"name"`
}

func TestMakeTransactionWithNested(t *testing.T) {
	transaction := NewTransaction(RequestContext{}, false)
	foos := []any{Foo{
		Entity: transaction.MakeEntity("foo"),
		Bars: []Bar{{
			Entity: transaction.MakeEntity("bar"),
			Name:   "Murphy's",
		}, {
			Entity: transaction.MakeEntity("bar"),
			Name:   "Irish Pub",
		}},
		Refs: ManyRef{Add: []string{"foo", "bar"}},
	}}
	transactionEntity, err := makeTransaction(foos, "")
	if err != nil {
		t.Failed()
	}
	if len(transactionEntity.Data) != 3 {
		t.Errorf("Incorrect number of entities in transaction")
	}
}
