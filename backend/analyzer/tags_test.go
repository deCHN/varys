package analyzer

import (
	"context"
	"testing"
)

func TestTagSanitization(t *testing.T) {
	// 1. Mock
	mock := &MockProvider{
		Response: `{
			"summary": "Summary",
			"key_points": [],
			"tags": ["note taking", "product review", "software-dev"],
			"assessment": {}
		}`,
	}

	// 2. Init
	an := &Analyzer{provider: mock}

	// 3. Run
	result, err := an.Analyze(context.Background(), "content", "", "English", 4096, nil)
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
