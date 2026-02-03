package analyzer

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

//go:embed default_prompt.txt
var defaultAnalysisPrompt string

type Analyzer struct {
	modelName string
	apiURL    string
}

func NewAnalyzer(model string) *Analyzer {
	if model == "" {
		model = "qwen3:8b" // Default
	}
	return &Analyzer{
		modelName: model,
		apiURL:    "http://localhost:11434/api/generate",
	}
}

type Request struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type Response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type AnalysisResult struct {
	Summary    string            `json:"summary"`
	KeyPoints  []string          `json:"key_points"`
	Tags       []string          `json:"tags"`
	Assessment map[string]string `json:"assessment"`
}

func (a *Analyzer) Analyze(text string, customPrompt string, targetLang string, contextSize int, onToken func(string)) (*AnalysisResult, error) {
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}
	if contextSize == 0 {
		contextSize = 8192 // Fallback default
	}

	var prompt string
	if customPrompt != "" {
		prompt = fmt.Sprintf("%s\n\nText to analyze:\n%s", customPrompt, text)
	} else {
		prompt = fmt.Sprintf(defaultAnalysisPrompt, targetLang, targetLang, targetLang, text)
	}

	reqBody := Request{
		Model:  a.modelName,
		Prompt: prompt,
		Stream: true,
		Options: map[string]interface{}{
			"num_ctx": contextSize,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(a.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %s: %s", resp.Status, string(body))
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var result Response
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode stream: %w", err)
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
	if idx := strings.Index(responseText, "{"); idx != -1 {
		responseText = responseText[idx:]
	}
	if idx := strings.LastIndex(responseText, "}"); idx != -1 {
		responseText = responseText[:idx+1]
	}

	var analysis AnalysisResult
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		// Fallback: try to return just summary if JSON fails
		return &AnalysisResult{Summary: responseText}, nil
	}

	// Post-process: Sanitize Tags
	for i, tag := range analysis.Tags {
		// Replace spaces with hyphens
		analysis.Tags[i] = strings.ReplaceAll(tag, " ", "-")
	}

	return &analysis, nil
}
