package integration_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"Varys/backend/translation"
)

// TestIntegrationTranslate performs a real call to Ollama.
// It requires Ollama to be running with the qwen3:0.6b model (or configured model).
func TestIntegrationTranslate(t *testing.T) {
	// 1. Check if Ollama is running
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		t.Skip("Ollama not running, skipping integration test")
	}
	defer resp.Body.Close()

	// 2. Init with default small model
	tr := translation.NewTranslator("qwen3:0.6b")

	// 3. Run Real Translation
	// "Hello" -> Simplified Chinese
	start := time.Now()
	results, err := tr.Translate("Hello", "Simplified Chinese", 4096, nil)
	if err != nil {
		t.Fatalf("Real translation failed: %v", err)
	}
	duration := time.Since(start)

	t.Logf("Translation took %v", duration)

	if len(results) == 0 {
		t.Fatal("No translation results returned")
	}

	translatedText := results[0].Translated
	t.Logf("Translated: %s", translatedText)

	// 4. Verify
	// Common translations for "Hello" in Simplified Chinese
	valid := []string{"你好", "您好", "哈喽", "喂"}
	found := false
	for _, v := range valid {
		if strings.Contains(translatedText, v) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected translation to contain one of %v, got: %s", valid, translatedText)
	}
}
