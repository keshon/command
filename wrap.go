package command

import "context"

const maxUnwrapDepth = 256

// Unwrappable is implemented by wrapped commands so adapters can reach the
// inner command (e.g. to type-assert to SlashProvider or ComponentHandler).
type Unwrappable interface {
	Command
	Unwrap() Command
}

// Wrapped wraps a command with a custom Run function. Used by middleware; the
// inner command is exposed via Unwrap() so adapters can access provider interfaces.
// Middleware should use Wrap to construct wrapped commands.
type Wrapped struct {
	Inner   Command
	RunFunc func(ctx context.Context, inv *Invocation) error
}

// Name delegates to the inner command.
func (w *Wrapped) Name() string { return w.Inner.Name() }

// Description delegates to the inner command.
func (w *Wrapped) Description() string { return w.Inner.Description() }

// Run executes the wrapper's RunFunc, or delegates to the inner command if RunFunc is nil.
func (w *Wrapped) Run(ctx context.Context, inv *Invocation) error {
	if w.RunFunc != nil {
		return w.RunFunc(ctx, inv)
	}
	return w.Inner.Run(ctx, inv)
}

// Unwrap returns the inner command.
func (w *Wrapped) Unwrap() Command { return w.Inner }

// Wrap returns a command that executes run instead of c.Run, delegating Name and
// Description to c. Use from middleware; the returned command implements Unwrappable.
//
// Wrap panics if c is nil.
func Wrap(c Command, run func(ctx context.Context, inv *Invocation) error) Command {
	if c == nil {
		panic("command: Wrap(nil, ...)")
	}
	return &Wrapped{Inner: c, RunFunc: run}
}

// Root returns the innermost command by repeatedly unwrapping until the command
// does not implement Unwrappable. Root returns nil if c is nil or if Unwrap
// returns nil. Unwrapping stops after maxUnwrapDepth steps to avoid infinite
// loops when Unwrap forms a cycle; the command at that depth is returned.
func Root(c Command) Command {
	if c == nil {
		return nil
	}
	for i := 0; i < maxUnwrapDepth; i++ {
		u, ok := c.(Unwrappable)
		if !ok {
			return c
		}
		next := u.Unwrap()
		if next == nil {
			return nil
		}
		c = next
	}
	return c
}
