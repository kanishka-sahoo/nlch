// Package plugin defines the Plugin interface and registry for context plugins.
package plugin

import (
	"github.com/kanishka-sahoo/nlch/internal/context"
)

// Plugin is the interface for context plugins.
type Plugin interface {
	Name() string
	Gather(ctx *context.Context) error
}

// Registry holds registered plugins.
var registry = make(map[string]Plugin)

// Register adds a plugin to the registry.
func Register(p Plugin) {
	registry[p.Name()] = p
}

// Get returns a plugin by name.
func Get(name string) (Plugin, bool) {
	p, ok := registry[name]
	return p, ok
}

// List returns all registered plugins.
func List() []Plugin {
	plugins := make([]Plugin, 0, len(registry))
	for _, p := range registry {
		plugins = append(plugins, p)
	}
	return plugins
}
