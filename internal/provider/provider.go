// Package provider defines the Provider interface and registry for LLM backends.
package provider

import (
	"github.com/kanishka-sahoo/nlch/internal/config"
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

// RegisterProvidersFromConfig registers all configured providers
func RegisterProvidersFromConfig(configProviders map[string]config.ProviderConfig) {
	for name, providerConfig := range configProviders {
		switch name {
		case "openrouter":
			if providerConfig.Key != "" {
				Register(&OpenRouterProvider{
					APIKey: providerConfig.Key,
					Model:  providerConfig.DefaultModel,
				})
			}
		case "anthropic":
			if providerConfig.Key != "" {
				Register(&AnthropicProvider{
					APIKey: providerConfig.Key,
					Model:  providerConfig.DefaultModel,
				})
			}
		case "openai":
			if providerConfig.Key != "" {
				Register(&OpenAIProvider{
					APIKey: providerConfig.Key,
					Model:  providerConfig.DefaultModel,
				})
			}
		case "gemini":
			if providerConfig.Key != "" {
				Register(&GeminiProvider{
					APIKey: providerConfig.Key,
					Model:  providerConfig.DefaultModel,
				})
			}
		case "ollama":
			url := providerConfig.URL
			if url == "" {
				url = "http://localhost:11434"
			}
			Register(&OllamaProvider{
				URL:   url,
				Model: providerConfig.DefaultModel,
			})
		}
	}
}
