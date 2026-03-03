package translation

import (
	"context"
	"strings"
	"testing"
)

// MockProvider implements analyzer.LLMProvider for testing
type MockProvider struct {
	Response string
}

func (m *MockProvider) Chat(ctx context.Context, prompt string, opts map[string]interface{}, cb func(string)) (string, error) {
	return m.Response, nil
}
func (m *MockProvider) Name() string { return "mock" }
func (m *MockProvider) Model() string { return "mock-model" }
func (m *MockProvider) ListModels(ctx context.Context) ([]string, error) { return []string{"mock-model"}, nil }

func TestTranslate(t *testing.T) {
	// 1. Init translator with mock provider
	mock := &MockProvider{
		Response: "1. Hello.\n2. World.",
	}
	tr := NewTranslator(mock)

	// 2. Run
	input := "你好。\n世界。"
	// Translate method now requires context as the first argument
	results, err := tr.Translate(context.Background(), input, "English", 4096, nil)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if !strings.Contains(results[0].Translated, "Hello") {
		t.Errorf("Expected 'Hello', got: %s", results[0].Translated)
	}

	if !strings.Contains(results[1].Translated, "World") {
		t.Errorf("Expected 'World', got: %s", results[1].Translated)
	}
}
