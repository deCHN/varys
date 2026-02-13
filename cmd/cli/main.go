package main

import (
	"Varys/backend/config"
	"Varys/backend/dependency"
	"Varys/backend/service"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// CLIPresenter implements service.EventLogger for terminal output.
type CLIPresenter struct{}

func (p *CLIPresenter) Log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), msg)
}

func (p *CLIPresenter) Progress(percent float64) {
	fmt.Printf("\rProgress: [%-50s] %.1f%%", strings.Repeat("=", int(percent/2)), percent)
	if percent >= 100 {
		fmt.Println()
	}
}

func (p *CLIPresenter) AnalysisChunk(token string) {
	fmt.Print(token)
}

func (p *CLIPresenter) Error(err error) {
	fmt.Fprintf(os.Stderr, "\n[ERROR] %v\n", err)
}

func main() {
	url := flag.String("url", "", "URL or local path to process")
	audioOnly := flag.Bool("audio-only", true, "Extract audio only")
	flag.Parse()

	if *url == "" {
		fmt.Println("Usage: varys-cli --url <URL or local path>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 1. Init Dependencies
	dm, err := dependency.NewManager()
	if err != nil {
		fmt.Printf("Dependency Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Load Config
	cm, _ := config.NewManager()
	cfg, err := cm.Load()
	if err != nil {
		fmt.Printf("Config Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Init Service
	svc := service.NewCoreService(dm)
	presenter := &CLIPresenter{}

	opts := service.Options{
		AudioOnly:      *audioOnly,
		ModelPath:      cfg.ModelPath,
		LLMModel:       cfg.LLMModel,
		TranslationMod: cfg.TranslationModel,
		AIProvider:     cfg.AIProvider,
		OpenAIKey:      cfg.OpenAIKey,
		OpenAIModel:    cfg.OpenAIModel,
		TargetLanguage: cfg.TargetLanguage,
		ContextSize:    cfg.ContextSize,
		CustomPrompt:   cfg.CustomPrompt,
		VaultPath:      cfg.VaultPath,
	}

	if opts.ContextSize == 0 {
		opts.ContextSize = 8192
	}

	// 4. Run Task
	fmt.Printf("Varys CLI starting task: %s\n", *url)
	result, err := svc.ProcessTask(context.Background(), *url, opts, presenter)
	if err != nil {
		fmt.Printf("\nTask failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccess! Note saved to: %s\n", result.NotePath)
}