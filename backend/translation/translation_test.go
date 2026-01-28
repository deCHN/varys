package translation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranslate(t *testing.T) {
	// 1. Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Return mock response (JSON array of pairs)
		resp := Response{
			Response: `[{"original": "Hello", "translated": "你好"}]`,
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
	results, err := tr.Translate("Hello", "Simplified Chinese", 4096, nil)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Translated != "你好" {
		t.Errorf("Unexpected result: %s", results[0].Translated)
	}
}
