package command_test

import (
	"context"
	"fmt"

	"github.com/keshon/command"
)

type pingCommand struct{}

func (pingCommand) Name() string        { return "ping" }
func (pingCommand) Description() string { return "Health check" }
func (pingCommand) Run(context.Context, *command.Invocation) error {
	return nil
}

func ExampleRegistry_Register() {
	r := command.NewRegistry()
	r.Register(pingCommand{})

	cmd := r.Get("ping")
	fmt.Println(cmd.Name())
	// Output: ping
}

func ExampleApply() {
	var log []string
	cmd := loggingCommand{log: &log}
	wrapped := command.Apply(cmd, labelMiddleware("outer", &log), labelMiddleware("inner", &log))
	_ = wrapped.Run(context.Background(), &command.Invocation{})

	fmt.Println(log[0], log[2])
	// Output: outer:before run
}

type loggingCommand struct {
	log *[]string
}

func (c loggingCommand) Name() string        { return "log" }
func (c loggingCommand) Description() string { return "log" }
func (c loggingCommand) Run(context.Context, *command.Invocation) error {
	*c.log = append(*c.log, "run")
	return nil
}

func labelMiddleware(label string, log *[]string) command.Middleware {
	return func(next command.Command) command.Command {
		return command.Wrap(next, func(ctx context.Context, inv *command.Invocation) error {
			*log = append(*log, label+":before")
			err := next.Run(ctx, inv)
			*log = append(*log, label+":after")
			return err
		})
	}
}

type slashProvider interface {
	SlashName() string
}

type slashPing struct {
	pingCommand
}

func (slashPing) SlashName() string { return "slash-ping" }

func ExampleRoot() {
	inner := slashPing{}
	wrapped := command.Wrap(inner, func(ctx context.Context, inv *command.Invocation) error {
		return inner.Run(ctx, inv)
	})

	root := command.Root(wrapped)
	provider := root.(slashProvider)
	fmt.Println(provider.SlashName())
	// Output: slash-ping
}
