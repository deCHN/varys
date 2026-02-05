package analyzer_test

import (
	"context"
	"os"
	"testing"
	"time"

	"Varys/backend/analyzer"
)

// TestIntegrationAnalyzeOpenAI tests the OpenAI provider with a real API call.
// Requires OPENAI_API_KEY to be set.
func TestIntegrationAnalyzeOpenAI(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}

	// 1. Initialize Analyzer with OpenAI
	an := analyzer.NewAnalyzer("openai", apiKey, "gpt-4o")

	// 2. Sample Text
	text := "Go is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. It is syntactically similar to C, but with memory safety, garbage collection, structural typing, and CSP-style concurrency."

	// 3. Run Analysis
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Sending request to OpenAI...")
	result, err := an.Analyze(ctx, text, "", "English", 4096, func(token string) {
		// Optional: t.Log(token)
	})

	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 4. Verify Result
	if result.Summary == "" {
		t.Error("Summary is empty")
	}
	if len(result.KeyPoints) == 0 {
		t.Error("KeyPoints are empty")
	}
	
	t.Logf("OpenAI Analysis Result: %+v", result)
}

// TestIntegrationAnalyzeOllama tests the Ollama provider with a real API call.
// Requires Ollama running locally with the specified model.
func TestIntegrationAnalyzeOllama(t *testing.T) {
	model := "qwen3:8b"
	an := analyzer.NewAnalyzer("ollama", "", model)

	text := "Rust is a multi-paradigm, general-purpose programming language. Rust emphasizes performance, type safety, and concurrency."

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Log("Sending request to Ollama...")
	result, err := an.Analyze(ctx, text, "", "English", 4096, nil)
	
	if err != nil {
		t.Logf("Ollama analysis failed: %v", err)
		t.Skip("Skipping Ollama test (Ensure Ollama is running and model qwen3:8b is pulled)")
	}

	if result.Summary == "" {
		t.Error("Summary is empty")
	}
	
	t.Logf("Ollama Analysis Result: %+v", result)
}
