package zenmanage

import "testing"

func TestDefaultsCollectionCRUD(t *testing.T) {
	c := NewDefaultsCollection()
	if c.Size() != 0 {
		t.Fatalf("expected empty collection")
	}

	c.Set("feature-a", true)
	if !c.Has("feature-a") {
		t.Fatalf("expected key to exist")
	}
	v, ok := c.Get("feature-a")
	if !ok || v != true {
		t.Fatalf("expected bool value true")
	}

	c.Delete("feature-a")
	if c.Has("feature-a") {
		t.Fatalf("expected key to be removed")
	}
}

func TestDefaultsFromMap(t *testing.T) {
	c := DefaultsFromMap(map[string]any{"a": 1, "b": "x"})
	if c.Size() != 2 {
		t.Fatalf("expected size 2")
	}
	all := c.All()
	all["a"] = 99
	v, _ := c.Get("a")
	if v == 99 {
		t.Fatalf("expected All() to return copy")
	}
}
