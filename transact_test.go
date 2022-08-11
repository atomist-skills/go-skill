package skill

import "testing"

type testEntity struct {
	Entity
	Foo string
}

func TestMakeWithoutId(t *testing.T) {
	transaction := NewTransaction()
	entity := transaction.Make("foo")
	if entity.Entity == "" {
		t.Errorf("Expected entity to be set")
	}
	if entity.Entity[0:1] != "$" {
		t.Errorf("Expected entity should start with $")
	}
}

func TestMakeWithId(t *testing.T) {
	transaction := NewTransaction()
	entity := transaction.Make("foo", "$repo")
	if entity.Entity != "$repo" {
		t.Errorf("Expected entity to be set to $repo")
	}
}
