package analyzer

import "context"

// LLMProvider defines the interface for AI backends (Ollama, OpenAI, etc.)
type LLMProvider interface {
	// Chat sends a prompt to the LLM and returns the full response.
	// If streamCallback is provided, it will receive chunks of the response.
	Chat(ctx context.Context, prompt string, options map[string]interface{}, streamCallback func(string)) (string, error)
}
