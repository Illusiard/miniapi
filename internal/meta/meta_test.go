package meta

import "testing"

func TestRegistry_Entities_Copy(t *testing.T) {
	r := New()

	r.AddEntity(Entity{Name: "A", Table: "a", Module: "m"})
	r.AddEntity(Entity{Name: "B", Table: "b", Module: "m"})

	e1 := r.Entities()
	if len(e1) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(e1))
	}

	e1[0].Name = "HACK"
	e2 := r.Entities()

	if e2[0].Name != "A" {
		t.Fatalf("expected registry to be immutable from outside, got %q", e2[0].Name)
	}
}

func TestRegistry_Modules_Copy(t *testing.T) {
	r := New()

	r.AddModule(Module{Name: "ping", WithStore: false, Version: "0.1.0"})
	r.AddModule(Module{Name: "notes", WithStore: true, Version: "0.1.0"})

	m1 := r.Modules()
	if len(m1) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(m1))
	}

	m1[0].Name = "HACK"
	m2 := r.Modules()

	if m2[0].Name != "ping" {
		t.Fatalf("expected registry to be immutable from outside, got %q", m2[0].Name)
	}
}
