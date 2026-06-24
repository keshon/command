package command

import (
	"context"
	"testing"
)

type innerCommand struct {
	name        string
	description string
	runCalled   *bool
}

func (c innerCommand) Name() string        { return c.name }
func (c innerCommand) Description() string { return c.description }
func (c innerCommand) Run(ctx context.Context, inv *Invocation) error {
	*c.runCalled = true
	return nil
}

type slashProvider interface {
	SlashName() string
}

type slashCommand struct {
	innerCommand
}

func (c slashCommand) SlashName() string { return "slash-" + c.name }

func TestWrap_DelegatesIdentity(t *testing.T) {
	runCalled := false
	inner := innerCommand{name: "ping", description: "Health check", runCalled: &runCalled}
	wrapped := Wrap(inner, func(ctx context.Context, inv *Invocation) error {
		return nil
	})

	if wrapped.Name() != "ping" {
		t.Errorf("Name() = %q, want %q", wrapped.Name(), "ping")
	}
	if wrapped.Description() != "Health check" {
		t.Errorf("Description() = %q, want %q", wrapped.Description(), "Health check")
	}
}

func TestWrap_NilRunFuncDelegatesToInner(t *testing.T) {
	runCalled := false
	inner := innerCommand{name: "ping", description: "desc", runCalled: &runCalled}
	wrapped := &Wrapped{Inner: inner, RunFunc: nil}

	if err := wrapped.Run(context.Background(), &Invocation{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !runCalled {
		t.Error("inner Run was not called")
	}
}

func TestWrap_PanicsOnNilCommand(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Wrap(nil, ...) did not panic")
		}
	}()
	Wrap(nil, func(ctx context.Context, inv *Invocation) error { return nil })
}

func TestRoot_UnwrapsMultipleLevels(t *testing.T) {
	runCalled := false
	inner := innerCommand{name: "ping", description: "desc", runCalled: &runCalled}

	level1 := Wrap(inner, func(ctx context.Context, inv *Invocation) error {
		return inner.Run(ctx, inv)
	})
	level2 := Wrap(level1, func(ctx context.Context, inv *Invocation) error {
		return level1.Run(ctx, inv)
	})

	root := Root(level2)
	if root != inner {
		t.Error("Root did not return innermost command")
	}
}

func TestRoot_TypeAssertToProviderInterface(t *testing.T) {
	runCalled := false
	inner := slashCommand{
		innerCommand: innerCommand{name: "ping", description: "desc", runCalled: &runCalled},
	}
	wrapped := Wrap(inner, func(ctx context.Context, inv *Invocation) error {
		return inner.Run(ctx, inv)
	})

	root := Root(wrapped)
	provider, ok := root.(slashProvider)
	if !ok {
		t.Fatal("Root command does not implement slashProvider")
	}
	if provider.SlashName() != "slash-ping" {
		t.Errorf("SlashName() = %q, want %q", provider.SlashName(), "slash-ping")
	}
}

func TestRoot_NilInput(t *testing.T) {
	if got := Root(nil); got != nil {
		t.Errorf("Root(nil) = %v, want nil", got)
	}
}

type nilUnwrappable struct {
	innerCommand
}

func (c nilUnwrappable) Unwrap() Command { return nil }

func TestRoot_UnwrapReturnsNil(t *testing.T) {
	runCalled := false
	inner := nilUnwrappable{
		innerCommand: innerCommand{name: "ping", description: "desc", runCalled: &runCalled},
	}
	if got := Root(inner); got != nil {
		t.Errorf("Root with Unwrap() == nil = %v, want nil", got)
	}
}

type cyclicCommand struct {
	innerCommand
	self Command
}

func (c *cyclicCommand) Unwrap() Command { return c.self }

func TestRoot_MaxDepth(t *testing.T) {
	runCalled := false
	inner := innerCommand{name: "ping", description: "desc", runCalled: &runCalled}
	cycle := &cyclicCommand{
		innerCommand: inner,
	}
	cycle.self = cycle

	if got := Root(cycle); got == nil {
		t.Fatal("Root returned nil for cyclic command")
	}

	long := Command(inner)
	for range maxUnwrapDepth + 10 {
		long = Wrap(long, func(ctx context.Context, inv *Invocation) error {
			return nil
		})
	}
	got := Root(long)
	if got == nil {
		t.Fatal("Root returned nil for deep chain")
	}
	if got.Name() != "ping" {
		t.Errorf("Name() = %q, want %q", got.Name(), "ping")
	}
}
