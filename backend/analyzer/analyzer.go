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

// GetDefaultPrompt returns the raw embedded prompt template.
func GetDefaultPrompt() string {
	return defaultAnalysisPrompt
}

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
	Provider   string            `json:"provider"`
	Model      string            `json:"model"`
}

// RenderPrompt applies placeholders to the prompt template.
func RenderPrompt(template string, targetLang string, text string, isCustom bool) string {
	if isCustom {
		// For custom prompts, we only replace {{.Content}} if present, otherwise append it.
		if strings.Contains(template, "{{.Content}}") {
			return strings.ReplaceAll(template, "{{.Content}}", text)
		}
		return fmt.Sprintf("%s\n\nText to analyze:\n%s", template, text)
	}

	// For the default prompt, replace both Language and Content placeholders
	rendered := strings.ReplaceAll(template, "{{.Language}}", targetLang)
	return strings.ReplaceAll(rendered, "{{.Content}}", text)
}

func (a *Analyzer) Analyze(ctx context.Context, text string, customPrompt string, targetLang string, contextSize int, onToken func(string)) (*AnalysisResult, error) {
	if targetLang == "" {
		targetLang = "English"
	}

	var prompt string
	if customPrompt != "" {
		prompt = RenderPrompt(customPrompt, targetLang, text, true)
	} else {
		prompt = RenderPrompt(defaultAnalysisPrompt, targetLang, text, false)
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
		analysis = AnalysisResult{Summary: responseText}
	}

	// Fill provider info
	analysis.Provider = a.provider.Name()
	analysis.Model = a.provider.Model()

	// Post-process: Sanitize Tags
	for i, tag := range analysis.Tags {
		// Replace spaces with hyphens
		analysis.Tags[i] = strings.ReplaceAll(tag, " ", "-")
	}

	return &analysis, nil
}

func (a *Analyzer) ListModels(ctx context.Context) ([]string, error) {
	return a.provider.ListModels(ctx)
}

func (a *Analyzer) GetProvider() LLMProvider {
	return a.provider
}
