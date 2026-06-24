// Package command provides a transport-agnostic command core. A command has
// a name, description, and Run(ctx, invocation). Registration and dispatch
// (Discord slash, CLI, HTTP) are defined by adapters that use this package.
//
// The package models a flat command registry: subcommands, groups, and aliases
// belong in adapters, not here.
//
// Registry is safe for concurrent Register, Get, and GetAll. Register and Wrap
// panic on programmer errors (nil command, empty name).
//
// Invocation.Data is an opaque adapter-defined payload; nil is valid and its
// meaning is defined by the adapter.
package command

import "context"

// Invocation carries the minimal input a command runner passes: arguments and
// an opaque payload. Adapters set Data to their context (e.g. *discordgo.Session
// and event, or *flag.FlagSet and CLI context).
type Invocation struct {
	Args []string
	Data interface{}
}

// Command is the core contract: identity and execution. Permissions, flags,
// subcommands, and transport-specific registration belong in adapters.
type Command interface {
	Name() string
	Description() string
	Run(ctx context.Context, inv *Invocation) error
}
