// Package provider implements the Anthropic Claude provider.
package provider

import (
	"github.com/kanishka-sahoo/nlch/internal/context"
)

type AnthropicProvider struct {
	BaseHTTPProvider
}

func (a *AnthropicProvider) Name() string { return "anthropic" }

func (a *AnthropicProvider) GetEndpoint() string {
	return "https://api.anthropic.com/v1/messages"
}

func (a *AnthropicProvider) GetHeaders(apiKey string) map[string]string {
	return map[string]string{
		"X-API-Key":         apiKey,
		"Content-Type":      "application/json",
		"anthropic-version": "2023-06-01",
	}
}

func (a *AnthropicProvider) BuildRequestBody(model, prompt string) ([]byte, error) {
	return BuildAnthropicRequestBody(model, prompt)
}

func (a *AnthropicProvider) ParseResponse(body []byte) (string, error) {
	return ParseAnthropicResponse(body)
}

func (a *AnthropicProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := a.Model
	if opts.Model != "" {
		model = opts.Model
	}

	return a.MakeHTTPRequest(a, model, promptStr)
}
