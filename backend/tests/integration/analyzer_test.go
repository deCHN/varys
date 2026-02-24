//go:build integration
package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"Varys/backend/analyzer"
)

func TestIntegrationAnalysis(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	tests := []struct {
		name     string
		provider string
		model    string
		skip     bool
	}{
		{
			name:     "OpenAI GPT-4o",
			provider: "openai",
			model:    "gpt-4o",
			skip:     apiKey == "",
		},
		{
			name:     "OpenAI GPT-4o-mini",
			provider: "openai",
			model:    "gpt-4o-mini",
			skip:     apiKey == "",
		},
		{
			name:     "Ollama Qwen3:8b (Default)",
			provider: "ollama",
			model:    "qwen3:8b",
			skip:     false,
		},
	}

	text := "Go is a statically typed, compiled programming language designed at Google. It is syntactically similar to C, but with memory safety, garbage collection, structural typing, and CSP-style concurrency."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping test due to missing config")
			}

			// Initialize Analyzer
			key := ""
			if tt.provider == "openai" {
				key = apiKey
			}
			an := analyzer.NewAnalyzer(tt.provider, key, tt.model)

			// Timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			t.Logf("[START] Provider: %s, Model: %s", tt.provider, tt.model)
			start := time.Now()
			result, err := an.Analyze(ctx, text, "", "English", 4096, nil)
			duration := time.Since(start)

			if err != nil {
				if tt.provider == "ollama" {
					t.Logf("[SKIP] Ollama analysis failed/timed out after %v: %v", duration, err)
					t.Skip("Skipping Ollama test (Ensure Ollama is running and model is pulled)")
				} else {
					t.Fatalf("[FAIL] %s Analyze failed after %v: %v", tt.name, duration, err)
				}
			}

			// Verify Result
			if result.Summary == "" {
				t.Error("Summary is empty")
			}
			if len(result.KeyPoints) == 0 {
				t.Error("KeyPoints are empty")
			}

			// Verify Provider Metadata
			if result.Provider != tt.provider {
				t.Errorf("Expected Provider %s, got %s", tt.provider, result.Provider)
			}
			if result.Model != tt.model {
				t.Errorf("Expected Model %s, got %s", tt.model, result.Model)
			}

			t.Logf("[DONE] %s finished in %v", tt.name, duration)
		})
	}
}
