package translation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTranslate(t *testing.T) {
	// 1. Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify request body contains text prompt, not JSON format instruction
		// (Optional deep check)

		// Return mock response (Numbered text lines)
		resp := Response{
			Response: "1. 你好.\n2. 世界。",
			Done:     true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// 2. Init
	tr := NewTranslator("test-model")
	tr.apiURL = ts.URL // Override URL for testing

	// 3. Run
	// text, targetLang, contextSize, callback
	input := "Hello.\nWorld."
	results, err := tr.Translate(input, "Simplified Chinese", 4096, nil)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if strings.TrimRight(results[0].Translated, "。.") != "你好" {

		t.Errorf("Expected '你好', got: %s", results[0].Translated)

	}

	if strings.TrimRight(results[1].Translated, "。.") != "世界" {

		t.Errorf("Expected '世界', got: %s", results[1].Translated)

	}

}
