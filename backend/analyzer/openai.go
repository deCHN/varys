package analyzer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o"
	}
	
	config := openai.DefaultConfig(apiKey)
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	
	client := openai.NewClientWithConfig(config)
	return &OpenAIProvider{
		client: client,
		model:  model,
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, prompt string, options map[string]interface{}, streamCallback func(string)) (string, error) {
	if p.client == nil {
		return "", errors.New("openai client not initialized")
	}

	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: true,
	}
    
	if val, ok := options["temperature"]; ok {
		if t, ok := val.(float64); ok {
			req.Temperature = float32(t)
		} else if t, ok := val.(float32); ok {
			req.Temperature = t
		}
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai stream error: %w", err)
	}
	defer stream.Close()

	var fullResponse strings.Builder
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("stream recv error: %w", err)
		}

		content := response.Choices[0].Delta.Content
		fullResponse.WriteString(content)
		if streamCallback != nil {
			streamCallback(content)
		}
	}

	return fullResponse.String(), nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Model() string {
	return p.model
}