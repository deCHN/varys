package analyzer

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed default_prompt.txt
var defaultAnalysisPrompt string

type Analyzer struct {
	provider LLMProvider
}

func NewAnalyzer(providerType, apiKey, model string) *Analyzer {
	var provider LLMProvider
	if providerType == "openai" {
		provider = NewOpenAIProvider(apiKey, model)
	} else {
		provider = NewOllamaProvider(model)
	}
	return &Analyzer{
		provider: provider,
	}
}

type AnalysisResult struct {
	Summary    string            `json:"summary"`
	KeyPoints  []string          `json:"key_points"`
	Tags       []string          `json:"tags"`
	Assessment map[string]string `json:"assessment"`
}

func (a *Analyzer) Analyze(ctx context.Context, text string, customPrompt string, targetLang string, contextSize int, onToken func(string)) (*AnalysisResult, error) {
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}
	
	var prompt string
	if customPrompt != "" {
		prompt = fmt.Sprintf("%s\n\nText to analyze:\n%s", customPrompt, text)
	} else {
		prompt = fmt.Sprintf(defaultAnalysisPrompt, targetLang, targetLang, targetLang, text)
	}

	options := map[string]interface{}{
		"num_ctx":     contextSize,
		"temperature": 0.1,
	}

	responseText, err := a.provider.Chat(ctx, prompt, options, onToken)
	if err != nil {
		return nil, err
	}

	// Clean up markdown code blocks if the LLM wrapped the output
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