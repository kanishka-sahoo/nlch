// Package provider implements the OpenAI GPT provider.
package provider

import (
	"github.com/kanishka-sahoo/nlch/internal/context"
)

type OpenAIProvider struct {
	BaseHTTPProvider
}

func (o *OpenAIProvider) Name() string { return "openai" }

func (o *OpenAIProvider) GetEndpoint() string {
	return "https://api.openai.com/v1/chat/completions"
}

func (o *OpenAIProvider) GetHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/json",
	}
}

func (o *OpenAIProvider) BuildRequestBody(model, prompt string) ([]byte, error) {
	return BuildOpenAIStyleRequestBody(model, prompt)
}

func (o *OpenAIProvider) ParseResponse(body []byte) (string, error) {
	return ParseOpenAIStyleResponse(body)
}

func (o *OpenAIProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := o.Model
	if opts.Model != "" {
		model = opts.Model
	}

	return o.MakeHTTPRequest(o, model, promptStr)
}
