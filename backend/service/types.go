package service

import (
	"context"
)

// EventLogger defines the interface for reporting progress and logs from the core service.
type EventLogger interface {
	Log(message string)
	Progress(percentage float64)
	AnalysisChunk(token string)
	Error(err error)
}

// Options defines the configuration for a processing task.
type Options struct {
	AudioOnly      bool
	ModelPath      string
	LLMModel       string
	TranslationMod string
	AIProvider     string
	OpenAIKey      string
	OpenAIModel    string
	TargetLanguage string
	ContextSize    int
	CustomPrompt   string
	VaultPath      string
}

// TaskResult contains the output of a successful processing task.
type TaskResult struct {
	NotePath  string
	MediaFile string
	Title     string
}

// Processor defines the core logic for the Varys pipeline.
type Processor interface {
	ProcessTask(ctx context.Context, url string, opts Options, logger EventLogger) (*TaskResult, error)
}
