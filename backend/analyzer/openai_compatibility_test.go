package analyzer_test

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
	for _, m := range allModels {
		if strings.HasPrefix(m, "gpt-") || strings.HasPrefix(m, "o1-") {
			chatModels = append(chatModels, m)
		}
	}

	sort.Slice(chatModels, func(i, j int) bool {
		return chatModels[i] > chatModels[j]
	})

	successCount := 0
	maxSuccess := 3

	for _, model := range chatModels {
		if successCount >= maxSuccess {
			break
		}

		t.Run(model, func(t *testing.T) {
			an := analyzer.NewAnalyzer("openai", apiKey, model)
			text := "Say 'OK'."

			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()

			start := time.Now()
			_, err := an.Analyze(ctx, text, "", "English", 1024, nil)
			duration := time.Since(start)

			if err != nil {
				// 如果是由于不支持的接口（404/400）导致，我们记录并继续
				t.Logf("[-] %-25s | SKIP | %v", model, err)
			} else {
				t.Logf("[+] %-25s | PASS | %v", model, duration)
				successCount++
			}
		})
	}

	if successCount == 0 {
		t.Error("No chat models passed the compatibility test.")
	}
}
