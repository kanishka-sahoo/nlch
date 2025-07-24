// Package provider defines the Provider interface and registry for LLM backends.
package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// HTTPProvider represents a provider configuration for HTTP-based APIs
type HTTPProvider interface {
	GetEndpoint() string
	GetHeaders(apiKey string) map[string]string
	BuildRequestBody(model, prompt string) ([]byte, error)
	ParseResponse(body []byte) (string, error)
}

// BaseHTTPProvider provides common HTTP functionality for API-based providers
type BaseHTTPProvider struct {
	APIKey string
	Model  string
}

// MakeHTTPRequest performs the common HTTP request logic
func (b *BaseHTTPProvider) MakeHTTPRequest(httpProvider HTTPProvider, model, prompt string) (string, error) {
	// Build request body
	reqBody, err := httpProvider.BuildRequestBody(model, prompt)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", httpProvider.GetEndpoint(), bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	// Set headers
	headers := httpProvider.GetHeaders(b.APIKey)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse response
	content, err := httpProvider.ParseResponse(body)
	if err != nil {
		return "", err
	}

	if content == "" {
		return "", errors.New("no content returned from API")
	}

	// Extract the first line as the shell command
	content = strings.TrimSpace(content)
	cmd := strings.SplitN(content, "\n", 2)[0]
	return cmd, nil
}

// BuildOpenAIStyleRequestBody creates an OpenAI-compatible request body
func BuildOpenAIStyleRequestBody(model, prompt string) ([]byte, error) {
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that generates safe, concise shell commands for the user's request."},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  128,
		"temperature": 0.2,
	}
	return json.Marshal(reqBody)
}

// ParseOpenAIStyleResponse parses an OpenAI-compatible response
func ParseOpenAIStyleResponse(body []byte) (string, error) {
	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if len(res.Choices) == 0 {
		return "", errors.New("no choices returned from API")
	}

	return res.Choices[0].Message.Content, nil
}

// BuildAnthropicRequestBody creates an Anthropic-specific request body
func BuildAnthropicRequestBody(model, prompt string) ([]byte, error) {
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 128,
		"system":     "You are a helpful assistant that generates safe, concise shell commands for the user's request.",
	}
	return json.Marshal(reqBody)
}

// ParseAnthropicResponse parses an Anthropic-specific response
func ParseAnthropicResponse(body []byte) (string, error) {
	var res struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if len(res.Content) == 0 {
		return "", errors.New("no content returned from API")
	}

	return res.Content[0].Text, nil
}

// BuildGeminiRequestBody creates a Gemini-specific request body
func BuildGeminiRequestBody(model, prompt string) ([]byte, error) {
	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": "You are a helpful assistant that generates safe, concise shell commands for the user's request.\n\n" + prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 128,
			"temperature":     0.2,
		},
	}
	return json.Marshal(reqBody)
}

// ParseGeminiResponse parses a Gemini-specific response
func ParseGeminiResponse(body []byte) (string, error) {
	var res struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content returned from API")
	}

	return res.Candidates[0].Content.Parts[0].Text, nil
}

// BuildOllamaRequestBody creates an Ollama-specific request body
func BuildOllamaRequestBody(model, prompt string) ([]byte, error) {
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that generates safe, concise shell commands for the user's request."},
			{"role": "user", "content": prompt},
		},
		"stream": false,
		"options": map[string]any{
			"num_predict": 128,
			"temperature": 0.2,
		},
	}
	return json.Marshal(reqBody)
}

// ParseOllamaResponse parses an Ollama-specific response
func ParseOllamaResponse(body []byte) (string, error) {
	var res struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if res.Message.Content == "" {
		return "", errors.New("no content returned from API")
	}

	return res.Message.Content, nil
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
					BaseHTTPProvider: BaseHTTPProvider{
						APIKey: providerConfig.Key,
						Model:  providerConfig.DefaultModel,
					},
				})
			}
		case "anthropic":
			if providerConfig.Key != "" {
				Register(&AnthropicProvider{
					BaseHTTPProvider: BaseHTTPProvider{
						APIKey: providerConfig.Key,
						Model:  providerConfig.DefaultModel,
					},
				})
			}
		case "openai":
			if providerConfig.Key != "" {
				Register(&OpenAIProvider{
					BaseHTTPProvider: BaseHTTPProvider{
						APIKey: providerConfig.Key,
						Model:  providerConfig.DefaultModel,
					},
				})
			}
		case "gemini":
			if providerConfig.Key != "" {
				Register(&GeminiProvider{
					BaseHTTPProvider: BaseHTTPProvider{
						APIKey: providerConfig.Key,
						Model:  providerConfig.DefaultModel,
					},
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
