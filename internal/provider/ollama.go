// Package provider implements the Ollama local LLM provider.
package provider

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kanishka-sahoo/nlch/internal/context"
)

type OllamaProvider struct {
	URL   string
	Model string
}

func (o *OllamaProvider) Name() string { return "ollama" }

func (o *OllamaProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := o.Model
	if opts.Model != "" {
		model = opts.Model
	}

	// Build request body
	reqBody, err := BuildOllamaRequestBody(model, promptStr)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", strings.TrimSuffix(o.URL, "/"))
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse response
	content, err := ParseOllamaResponse(body)
	if err != nil {
		return "", err
	}

	if content == "" {
		return "", errors.New("no content returned from Ollama")
	}

	// Extract the first line as the shell command
	content = strings.TrimSpace(content)
	cmd := strings.SplitN(content, "\n", 2)[0]
	return cmd, nil
}
