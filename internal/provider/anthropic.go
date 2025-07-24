// Package provider implements the Anthropic Claude provider.
package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kanishka-sahoo/nlch/internal/context"
)

type AnthropicProvider struct {
	APIKey string
	Model  string
}

func (a *AnthropicProvider) Name() string { return "anthropic" }

func (a *AnthropicProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := a.Model
	if opts.Model != "" {
		model = opts.Model
	}

	// Prepare Anthropic API request
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": promptStr},
		},
		"max_tokens": 128,
		"system":     "You are a helpful assistant that generates safe, concise shell commands for the user's request.",
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-API-Key", a.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error: %s", string(b))
	}

	var res struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if len(res.Content) == 0 {
		return "", errors.New("no content returned from Anthropic")
	}

	// Extract the first line as the shell command
	content := strings.TrimSpace(res.Content[0].Text)
	cmd := strings.SplitN(content, "\n", 2)[0]
	return cmd, nil
}
