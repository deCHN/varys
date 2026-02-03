package analyzer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyze(t *testing.T) {
	// 1. Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify body
		var req Request
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "test-model" {
			t.Errorf("Expected model test-model, got %s", req.Model)
		}
		if req.Prompt == "" {
			t.Error("Empty prompt")
		}

		// Return mock response
		resp := Response{
			Response: "This is a summary.",
			Done:     true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// 2. Init
	an := NewAnalyzer("test-model")
	an.apiURL = ts.URL // Override URL for testing

	// 3. Run
	// text, customPrompt, targetLang, contextSize, callback
	result, err := an.Analyze("Some text content", "", "English", 4096, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Summary != "This is a summary." {
		t.Errorf("Unexpected result: %s", result.Summary)
	}
}
