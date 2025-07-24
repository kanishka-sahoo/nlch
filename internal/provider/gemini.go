// Package provider implements the Google Gemini provider.
package provider

import (
	"fmt"

	"github.com/kanishka-sahoo/nlch/internal/context"
)

type GeminiProvider struct {
	BaseHTTPProvider
}

func (g *GeminiProvider) Name() string { return "gemini" }

func (g *GeminiProvider) GetEndpoint() string {
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.Model, g.APIKey)
}

func (g *GeminiProvider) GetHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func (g *GeminiProvider) BuildRequestBody(model, prompt string) ([]byte, error) {
	return BuildGeminiRequestBody(model, prompt)
}

func (g *GeminiProvider) ParseResponse(body []byte) (string, error) {
	return ParseGeminiResponse(body)
}

func (g *GeminiProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := g.Model
	if opts.Model != "" {
		model = opts.Model
	}

	return g.MakeHTTPRequest(g, model, promptStr)
}
