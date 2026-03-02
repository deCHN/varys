//go:build integration
package integration_test

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"Varys/backend/analyzer"
)

func TestOpenAIModelCompatibility(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	p := analyzer.NewOpenAIProvider(apiKey, "")
	ctx := context.Background()
	allModels, err := p.ListModels(ctx)
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	var chatModels []string
	// Predefined blacklist: these models are known to not support standard Chat API or require special parameters
	blacklist := []string{
		"realtime", "audio", "image", "tts", "search", 
		"transcribe", "pro", "codex", "instruct", "vision",
	}

	for _, m := range allModels {
		if !strings.HasPrefix(m, "gpt-") && !strings.HasPrefix(m, "o1-") {
			continue
		}

		isBlacklisted := false
		for _, b := range blacklist {
			if strings.Contains(strings.ToLower(m), b) {
				isBlacklisted = true
				break
			}
		}

		if !isBlacklisted {
			chatModels = append(chatModels, m)
		}
	}

	sort.Slice(chatModels, func(i, j int) bool {
		return chatModels[i] > chatModels[j]
	})

	successCount := 0
	maxSuccess := 3

	t.Logf("Candidate models after pre-filtering: %v", chatModels)

	for _, model := range chatModels {
		if successCount >= maxSuccess {
			break
		}

		// Only filtered models will execute t.Run (sending network requests)
		t.Run(model, func(t *testing.T) {
			an := analyzer.NewAnalyzer("openai", apiKey, model)
			text := "Say 'OK'."

			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()

			start := time.Now()
			_, err := an.Analyze(ctx, text, "", "English", 1024, nil)
			duration := time.Since(start)

			if err != nil {
				// If it still errors (e.g. some new models), log but don't count towards success
				t.Logf("[FAIL] %-25s | %v", model, err)
			} else {
				t.Logf("[PASS] %-25s | %v", model, duration)
				successCount++
			}
		})
	}

	if successCount == 0 {
		t.Error("No chat models passed the compatibility test.")
	}
}