package analyzer

import (
	"context"
	"testing"
)

type MockProvider struct {
	Response string
}

func (m *MockProvider) Chat(ctx context.Context, prompt string, options map[string]interface{}, streamCallback func(string)) (string, error) {
	if streamCallback != nil {
		streamCallback(m.Response)
	}
	return m.Response, nil
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) Model() string {
	return "test-model"
}

func (m *MockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"test-model"}, nil
}

func TestAnalyze(t *testing.T) {
	// 1. Mock LLM Response (Valid JSON)
	mock := &MockProvider{
		Response: `{
			"summary": "This is a summary.",
			"key_points": ["Point 1"],
			"tags": ["tag1"],
			"assessment": {}
		}`,
	}

	// 2. Init
	an := &Analyzer{provider: mock}

	// 3. Run
	result, err := an.Analyze(context.Background(), "Some text", "", "English", 4096, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Summary != "This is a summary." {
		t.Errorf("Unexpected result: %s", result.Summary)
	}
}