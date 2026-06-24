package command

import (
	"context"
	"testing"
)

type logCommand struct {
	log *[]string
}

func (c logCommand) Name() string        { return "log" }
func (c logCommand) Description() string { return "log" }
func (c logCommand) Run(ctx context.Context, inv *Invocation) error {
	*c.log = append(*c.log, "run")
	return nil
}

func loggingMiddleware(label string, log *[]string) Middleware {
	return func(next Command) Command {
		return Wrap(next, func(ctx context.Context, inv *Invocation) error {
			*log = append(*log, label+":before")
			err := next.Run(ctx, inv)
			*log = append(*log, label+":after")
			return err
		})
	}
}

func TestApply_NoMiddleware(t *testing.T) {
	var log []string
	cmd := logCommand{log: &log}
	wrapped := Apply(cmd)

	if wrapped.Name() != "log" {
		t.Errorf("Name() = %q, want %q", wrapped.Name(), "log")
	}
	if wrapped.Description() != "log" {
		t.Errorf("Description() = %q, want %q", wrapped.Description(), "log")
	}

	if err := wrapped.Run(context.Background(), &Invocation{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(log) != 1 || log[0] != "run" {
		t.Errorf("log = %v, want [run]", log)
	}
}

func TestApply_MiddlewareOrder(t *testing.T) {
	var log []string
	cmd := logCommand{log: &log}
	wrapped := Apply(cmd,
		loggingMiddleware("outer", &log),
		loggingMiddleware("inner", &log),
	)

	if err := wrapped.Run(context.Background(), &Invocation{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := []string{
		"outer:before",
		"inner:before",
		"run",
		"inner:after",
		"outer:after",
	}
	if len(log) != len(want) {
		t.Fatalf("log = %v, want %v", log, want)
	}
	for i, entry := range log {
		if entry != want[i] {
			t.Errorf("log[%d] = %q, want %q", i, entry, want[i])
		}
	}
}
