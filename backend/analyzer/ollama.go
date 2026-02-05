package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	modelName string
	apiURL    string
}

func NewOllamaProvider(model string) *OllamaProvider {
	if model == "" {
		model = "qwen3:8b"
	}
	return &OllamaProvider{
		modelName: model,
		apiURL:    "http://localhost:11434/api/generate",
	}
}

// Ollama API Structures
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (p *OllamaProvider) Chat(ctx context.Context, prompt string, options map[string]interface{}, streamCallback func(string)) (string, error) {
	reqBody := OllamaRequest{
		Model:   p.modelName,
		Prompt:  prompt,
		Stream:  true,
		Options: options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error %s: %s", resp.Status, string(body))
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var result OllamaResponse
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode stream: %w", err)
		}

		fullResponse.WriteString(result.Response)
		if streamCallback != nil {
			streamCallback(result.Response)
		}

		if result.Done {
			break
		}
	}

	return fullResponse.String(), nil
}
