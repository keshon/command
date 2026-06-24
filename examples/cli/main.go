package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/keshon/command"
)

type PingCommand struct{}

func (PingCommand) Name() string        { return "ping" }
func (PingCommand) Description() string { return "Health check" }
func (PingCommand) Run(context.Context, *command.Invocation) error {
	fmt.Println("pong")
	return nil
}

type EchoCommand struct{}

func (EchoCommand) Name() string        { return "echo" }
func (EchoCommand) Description() string { return "Print arguments" }
func (EchoCommand) Run(_ context.Context, inv *command.Invocation) error {
	fmt.Println(strings.Join(inv.Args, " "))
	return nil
}

func loggingMiddleware(next command.Command) command.Command {
	return command.Wrap(next, func(ctx context.Context, inv *command.Invocation) error {
		fmt.Printf("-> %s\n", next.Name())
		err := next.Run(ctx, inv)
		fmt.Printf("<- %s\n", next.Name())
		return err
	})
}

func main() {
	command.DefaultRegistry.Register(command.Apply(PingCommand{}, loggingMiddleware))
	command.DefaultRegistry.Register(command.Apply(EchoCommand{}, loggingMiddleware))

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	name := os.Args[1]
	args := os.Args[2:]

	cmd := command.DefaultRegistry.Get(name)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
		printUsage()
		os.Exit(1)
	}

	if err := cmd.Run(context.Background(), &command.Invocation{Args: args}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "available commands:")
	for _, cmd := range command.DefaultRegistry.GetAll() {
		fmt.Fprintf(os.Stderr, "  %s — %s\n", cmd.Name(), cmd.Description())
	}
}
