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

func TestNewAnalyzer(t *testing.T) {
	// Test Ollama (Default)
	an := NewAnalyzer("ollama", "", "qwen3:8b")
	if an.provider.Name() != "ollama" {
		t.Errorf("Expected ollama provider, got %s", an.provider.Name())
	}

	// Test OpenAI
	anOpenAI := NewAnalyzer("openai", "sk-test", "gpt-4o")
	if anOpenAI.provider.Name() != "openai" {
		t.Errorf("Expected openai provider, got %s", anOpenAI.provider.Name())
	}
}

func TestListModels(t *testing.T) {
	mock := &MockProvider{Response: "test"}
	an := &Analyzer{provider: mock}

	models, err := an.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 || models[0] != "test-model" {
		t.Errorf("Unexpected models: %v", models)
	}
}
