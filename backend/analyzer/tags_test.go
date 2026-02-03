package analyzer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTagSanitization(t *testing.T) {
	// 1. Mock Server returning tags with spaces
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock valid JSON response
		resp := Response{
			Response: `{
				"summary": "Summary",
				"key_points": [],
				"tags": ["note taking", "product review", "software-dev"],
				"assessment": {}
			}`,
			Done: true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// 2. Init
	an := NewAnalyzer("test-model")
	an.apiURL = ts.URL

	// 3. Run
	result, err := an.Analyze("content", "", "English", 4096, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 4. Verify Tags
	expectedTags := []string{"note-taking", "product-review", "software-dev"}

	if len(result.Tags) != len(expectedTags) {
		t.Fatalf("Expected %d tags, got %d", len(expectedTags), len(result.Tags))
	}

	for i, tag := range result.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Tag mismatch at index %d. Expected '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}
