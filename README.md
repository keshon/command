# command

A transport-agnostic command execution core for Go.

Defines a minimal contract for commands — identity + execution — and provides a registry, middleware, and safe adapter unwrapping. Transport concerns (CLI flags, HTTP routing, Discord slash definitions) belong in adapters, not here.

This library implements the **Ports & Adapters (Hexagonal Architecture)** pattern: `Command` is the port — a minimal, transport-agnostic contract — while CLI, Discord, HTTP, and other transports are adapters that live outside this package.

---

## When to use this

When you want the same command logic to run across multiple transports — CLI, Discord/Telegram bots, HTTP APIs, background workers — without coupling it to any of them.

The core models a **flat command registry**. Subcommands, groups, and aliases belong in adapters.

---

## Core concepts

### Command

```go
type Command interface {
    Name() string
    Description() string
    Run(ctx context.Context, inv *Invocation) error
}
```

Permissions, flags, subcommands, and routing are adapter concerns.

### Invocation

```go
type Invocation struct {
    Args []string
    Data interface{}
}
```

`Data` is an opaque adapter-defined payload (event, request, session). Type safety is enforced at the adapter boundary. `nil` is valid; meaning is defined by the adapter.

### Registry

Stores commands by name. Does not dispatch or execute — adapters decide how and when commands are invoked.

```go
command.DefaultRegistry.Register(cmd)
command.DefaultRegistry.Get("ping")
command.DefaultRegistry.GetAll()
```

`Registry` is safe for concurrent use: `Register` is exclusive; `Get` and `GetAll` may run concurrently with each other and with `Register`.

`Register` panics on programmer errors: `nil` command or empty `Name()`.

A global `DefaultRegistry` is provided for convenience; inject your own `Registry` where isolation matters (e.g. tests).

### Middleware

```go
type Middleware func(Command) Command
```

Logging, metrics, permission checks, panic recovery — anything cross-cutting. Stays transport-agnostic.

### Wrapping and unwrapping

`Wrap` replaces `Run` while preserving identity. Middleware should use `Wrap` so adapters can unwrap the chain.

`Root` unwraps a middleware chain back to the original command — useful when adapters need to type-assert to transport-specific interfaces (e.g. `SlashProvider`). `Root` returns `nil` for a `nil` input or when `Unwrap` returns `nil`. Unwrapping stops after a depth limit to avoid infinite loops on cyclic `Unwrap` chains.

---

## Usage

```go
// Define
type PingCommand struct{}
func (PingCommand) Name() string        { return "ping" }
func (PingCommand) Description() string { return "Health check" }
func (PingCommand) Run(ctx context.Context, inv *command.Invocation) error {
    return nil
}

// Register
command.DefaultRegistry.Register(PingCommand{})

// Execute (from an adapter)
cmd := command.DefaultRegistry.Get("ping")
err := cmd.Run(ctx, &command.Invocation{Args: args, Data: adapterCtx})
```

Install:

```bash
go get github.com/keshon/command
```

---

## Examples

A minimal CLI adapter lives in [`examples/cli`](examples/cli). It registers `ping` and `echo` commands, applies logging middleware, and dispatches by `os.Args`:

```bash
go run ./examples/cli ping
go run ./examples/cli echo hello world
```

Godoc examples are in the package test files (`ExampleRegistry_Register`, `ExampleApply`, `ExampleRoot`).

---

## Stability

This project is at **v0.x**. The API may change without a major version bump until v1.0. After v1.0, [semantic versioning](https://semver.org/) applies.

See [CHANGELOG.md](CHANGELOG.md) for release notes.

---

## What lives where

| Concern | command | Adapter |
|---|---|---|
| Identity | `Name()`, `Description()` | Category, permissions, help |
| Execution | `Run(ctx, *Invocation)` | Build `Invocation` from event/request |
| Registry | `Register`, `Get`, `GetAll` | Dispatch logic |
| Middleware | `Middleware`, `Apply`, `Wrap` | Transport-specific middleware |
| Registration | — | Slash definitions, flags, routes |
