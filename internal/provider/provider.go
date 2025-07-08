// Package provider defines the Provider interface and registry for LLM backends.
package provider

import (
	"github.com/kanishka-sahoo/nlch/internal/context"
)

// ProviderOptions holds options for provider calls (e.g., model override).
type ProviderOptions struct {
	Model    string
	Provider string
}

// Provider is the interface for LLM backends.
type Provider interface {
	Name() string
	GenerateCommand(ctx context.Context, prompt string, opts ProviderOptions) (string, error)
}

// Registry holds registered providers.
var registry = make(map[string]Provider)

// Register adds a provider to the registry.
func Register(p Provider) {
	registry[p.Name()] = p
}

// Get returns a provider by name.
func Get(name string) (Provider, bool) {
	p, ok := registry[name]
	return p, ok
}

// List returns all registered providers.
func List() []Provider {
	providers := make([]Provider, 0, len(registry))
	for _, p := range registry {
		providers = append(providers, p)
	}
	return providers
}
