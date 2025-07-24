// Package provider implements the Ollama local LLM provider.
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

	// Prepare Ollama API request
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that generates safe, concise shell commands for the user's request."},
			{"role": "user", "content": promptStr},
		},
		"stream": false,
		"options": map[string]any{
			"num_predict": 128,
			"temperature": 0.2,
		},
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/chat", strings.TrimSuffix(o.URL, "/"))
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error: %s", string(b))
	}

	var res struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.Message.Content == "" {
		return "", errors.New("no content returned from Ollama")
	}

	// Extract the first line as the shell command
	content := strings.TrimSpace(res.Message.Content)
	cmd := strings.SplitN(content, "\n", 2)[0]
	return cmd, nil
}
