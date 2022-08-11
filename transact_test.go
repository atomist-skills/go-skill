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
	transaction := NewTransaction()
	entity := transaction.MakeEntity("foo")
	if entity.Entity == "" {
		t.Errorf("Expected entity to be set")
	}
	if entity.Entity[0:1] != "$" {
		t.Errorf("Expected entity should start with $")
	}
}

func TestMakeWithId(t *testing.T) {
	transaction := NewTransaction()
	entity := transaction.MakeEntity("foo", "$repo")
	if entity.Entity != "$repo" {
		t.Errorf("Expected entity to be set to $repo")
	}
}

func TestAddEntities(t *testing.T) {
	transaction := NewTransaction()
	entity1 := testEntity{
		Entity: transaction.MakeEntity("foo"),
	}
	entity2 := testEntity{
		Entity: transaction.MakeEntity("bar"),
	}
	transaction.AddEntity(entity1)
	transaction.AddEntity(entity2)

	if len(transaction.Entities()) != 2 {
		t.Errorf("Expected two entities")
	}
	refs := transaction.EntityRefs("foo")
	if len(refs) != 1 {
		t.Errorf("Expected one entity ref")
	}
}
