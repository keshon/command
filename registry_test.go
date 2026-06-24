package command

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

type stubCommand struct {
	name string
	desc string
}

func (c stubCommand) Name() string                           { return c.name }
func (c stubCommand) Description() string                    { return c.desc }
func (c stubCommand) Run(context.Context, *Invocation) error { return nil }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	cmd := stubCommand{name: "ping", desc: "ping desc"}
	r.Register(cmd)

	got := r.Get("ping")
	if got == nil {
		t.Fatal("Get returned nil for registered command")
	}
	if got.Name() != "ping" {
		t.Errorf("got name %q, want %q", got.Name(), "ping")
	}
}

func TestRegistry_GetUnregistered(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("missing"); got != nil {
		t.Errorf("Get returned %v, want nil", got)
	}
}

func TestRegistry_RegisterOverwrites(t *testing.T) {
	r := NewRegistry()
	first := stubCommand{name: "ping", desc: "first"}
	second := stubCommand{name: "ping", desc: "second"}

	r.Register(first)
	r.Register(second)

	got := r.Get("ping")
	if got.Description() != "second" {
		t.Errorf("Description() = %q, want %q", got.Description(), "second")
	}
}

func TestRegistry_GetAllSorted(t *testing.T) {
	r := NewRegistry()
	r.Register(stubCommand{name: "echo", desc: "echo"})
	r.Register(stubCommand{name: "alpha", desc: "alpha"})
	r.Register(stubCommand{name: "ping", desc: "ping"})

	all := r.GetAll()
	if len(all) != 3 {
		t.Fatalf("GetAll returned %d commands, want 3", len(all))
	}

	names := []string{all[0].Name(), all[1].Name(), all[2].Name()}
	want := []string{"alpha", "echo", "ping"}
	for i, name := range names {
		if name != want[i] {
			t.Errorf("GetAll[%d] = %q, want %q", i, name, want[i])
		}
	}
}

func TestRegistry_ConcurrentRegisterGet(t *testing.T) {
	r := NewRegistry()
	const goroutines = 32
	const perGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range perGoroutine {
				name := fmt.Sprintf("cmd-%d-%d", id, i)
				r.Register(stubCommand{name: name, desc: name})
				if got := r.Get(name); got == nil {
					t.Errorf("Get(%q) returned nil after Register", name)
				}
				_ = r.GetAll()
			}
		}(g)
	}
	wg.Wait()
}

func TestRegister_PanicsOnNil(t *testing.T) {
	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("Register(nil) did not panic")
		}
	}()
	r.Register(nil)
}

func TestRegister_PanicsOnEmptyName(t *testing.T) {
	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("Register with empty name did not panic")
		}
	}()
	r.Register(stubCommand{name: "", desc: "no name"})
}
