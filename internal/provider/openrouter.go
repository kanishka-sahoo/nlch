// Package provider implements the OpenRouter LLM provider.
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

type OpenRouterProvider struct {
	APIKey string
	Model  string
}

func (o *OpenRouterProvider) Name() string { return "openrouter" }

func (o *OpenRouterProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := o.Model
	if opts.Model != "" {
		model = opts.Model
	}
	// Prepare OpenAI-compatible request
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that generates safe, concise shell commands for the user's request."},
			{"role": "user", "content": promptStr},
		},
		"max_tokens":  128,
		"temperature": 0.2,
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/kanishka-sahoo/nlch") // Optional, for OpenRouter compliance

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenRouter API error: %s", string(b))
	}
	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if len(res.Choices) == 0 {
		return "", errors.New("no choices returned from OpenRouter")
	}
	// Extract the first line as the shell command
	content := strings.TrimSpace(res.Choices[0].Message.Content)
	cmd := strings.SplitN(content, "\n", 2)[0]
	return cmd, nil
}

func init() {
	// Register OpenRouter provider if config is present
	cfg, err := config.Load()
	if err != nil {
		return
	}
	provCfg, ok := cfg.Providers["openrouter"]
	if !ok || provCfg.Key == "" {
		return
	}
	Register(&OpenRouterProvider{
		APIKey: provCfg.Key,
		Model:  provCfg.DefaultModel,
	})
}
