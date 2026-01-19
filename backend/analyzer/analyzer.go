package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Analyzer struct {
	modelName string
	apiURL    string
}

func NewAnalyzer(model string) *Analyzer {
	if model == "" {
		model = "qwen2.5:7b" // Default
	}
	return &Analyzer{
		modelName: model,
		apiURL:    "http://localhost:11434/api/generate",
	}
}

type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type Response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (a *Analyzer) Analyze(text string, onToken func(string)) (string, error) {
	// Simple analysis prompt
	prompt := fmt.Sprintf("Analyze the following text and provide a summary, key points, and tags in Simplified Chinese (zh-CN). Use Markdown format:\n\n%s", text)

	reqBody := Request{
		Model:  a.modelName,
		Prompt: prompt,
		Stream: true, // Enable streaming
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(a.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w. Is Ollama running?", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error %s: %s", resp.Status, string(body))
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var result Response
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode stream: %w", err)
		}

		fullResponse.WriteString(result.Response)
		if onToken != nil {
			onToken(result.Response)
		}

		if result.Done {
			break
		}
	}

	// Clean up markdown code blocks if the LLM wrapped the output
	responseText := fullResponse.String()
	// Remove starting ```markdown or ```
	if strings.HasPrefix(responseText, "```") {
		// Remove first line
		if idx := strings.Index(responseText, "\n"); idx != -1 {
			responseText = responseText[idx+1:]
		} else {
            // Just remove the marker if no newline (unlikely for block)
            responseText = strings.TrimPrefix(responseText, "```markdown")
            responseText = strings.TrimPrefix(responseText, "```")
        }
	}
	// Remove ending ```
	responseText = strings.TrimSuffix(strings.TrimSpace(responseText), "```")

	return responseText, nil
}
