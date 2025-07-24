// Package provider implements the OpenRouter LLM provider.
package provider

import (
	"github.com/kanishka-sahoo/nlch/internal/context"
)

type OpenRouterProvider struct {
	BaseHTTPProvider
}

func (o *OpenRouterProvider) Name() string { return "openrouter" }

func (o *OpenRouterProvider) GetEndpoint() string {
	return "https://openrouter.ai/api/v1/chat/completions"
}

func (o *OpenRouterProvider) GetHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/json",
		"HTTP-Referer":  "https://github.com/kanishka-sahoo/nlch", // Optional, for OpenRouter compliance
	}
}

func (o *OpenRouterProvider) BuildRequestBody(model, prompt string) ([]byte, error) {
	return BuildOpenAIStyleRequestBody(model, prompt)
}

func (o *OpenRouterProvider) ParseResponse(body []byte) (string, error) {
	return ParseOpenAIStyleResponse(body)
}

func (o *OpenRouterProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := o.Model
	if opts.Model != "" {
		model = opts.Model
	}

	return o.MakeHTTPRequest(o, model, promptStr)
}
