package command

import (
	"sort"
	"sync"
)

// DefaultRegistry is the global registry used by adapters (Discord, CLI, etc.).
// Use it for convenience; inject a separate *Registry where isolation is needed (e.g. tests).
//
// Safe for concurrent use: Register is exclusive; Get and GetAll may run
// concurrently with each other and with Register.
var DefaultRegistry = NewRegistry()

// Registry stores commands by name. It does not dispatch or execute; each adapter
// looks up commands and invokes them with its own context.
//
// Safe for concurrent use: Register is exclusive; Get and GetAll may run
// concurrently with each other and with Register.
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Command
}

// NewRegistry returns a new empty registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

// Register adds or replaces a command by name. Typically called from init() or adapter setup.
//
// Register panics if c is nil or c.Name() is empty.
func (r *Registry) Register(c Command) {
	if c == nil {
		panic("command: Register(nil)")
	}
	name := c.Name()
	if name == "" {
		panic("command: Register command with empty name")
	}

	r.mu.Lock()
	r.commands[name] = c
	r.mu.Unlock()
}

// Get returns the command with the given name, or nil if not registered.
func (r *Registry) Get(name string) Command {
	r.mu.RLock()
	c := r.commands[name]
	r.mu.RUnlock()
	return c
}

// GetAll returns all registered commands, sorted by name.
func (r *Registry) GetAll() []Command {
	r.mu.RLock()
	list := make([]Command, 0, len(r.commands))
	for _, c := range r.commands {
		list = append(list, c)
	}
	r.mu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}
