// Package provider implements the Google Gemini provider.
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

type GeminiProvider struct {
	APIKey string
	Model  string
}

func (g *GeminiProvider) Name() string { return "gemini" }

func (g *GeminiProvider) GenerateCommand(ctx context.Context, promptStr string, opts ProviderOptions) (string, error) {
	model := g.Model
	if opts.Model != "" {
		model = opts.Model
	}

	// Prepare Gemini API request
	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": "You are a helpful assistant that generates safe, concise shell commands for the user's request.\n\n" + promptStr},
				},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 128,
			"temperature":     0.2,
		},
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, g.APIKey)
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
		return "", fmt.Errorf("Gemini API error: %s", string(b))
	}

	var res struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content returned from Gemini")
	}

	// Extract the first line as the shell command
	content := strings.TrimSpace(res.Candidates[0].Content.Parts[0].Text)
	cmd := strings.SplitN(content, "\n", 2)[0]
	return cmd, nil
}
