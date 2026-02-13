package main

import (
	"Varys/backend/config"
	"Varys/backend/dependency"
	"Varys/backend/service"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
	var (
		audioOnly      bool
		aiProvider     string
		model          string
		translationMod string
		targetLang     string
		contextSize    int
		vaultPath      string
	)

	rootCmd := &cobra.Command{
		Use:   "varys-cli [URL or local path]",
		Short: "Varys CLI - Transcribe and analyze audio/video content",
		Long: `Varys is a local-first tool that extracts audio from URLs or local files, 
transcribes them using Whisper, and generates structured analysis using LLMs (Ollama/OpenAI).`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]

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

			// 3. Merge CLI Flags with Config
			opts := service.Options{
				AudioOnly:      audioOnly,
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

			// Override if flags are provided
			if cmd.Flags().Changed("ai-provider") {
				opts.AIProvider = aiProvider
			}
			if cmd.Flags().Changed("model") {
				if opts.AIProvider == "openai" {
					opts.OpenAIModel = model
				} else {
					opts.LLMModel = model
				}
			}
			if cmd.Flags().Changed("translation-model") {
				opts.TranslationMod = translationMod
			}
			if cmd.Flags().Changed("target-lang") {
				opts.TargetLanguage = targetLang
			}
			if cmd.Flags().Changed("context-size") {
				opts.ContextSize = contextSize
			}
			if cmd.Flags().Changed("vault") {
				opts.VaultPath = vaultPath
			}

			if opts.ContextSize == 0 {
				opts.ContextSize = 8192
			}

			// 4. Init Service
			svc := service.NewCoreService(dm)
			presenter := &CLIPresenter{}

			fmt.Printf("Varys CLI starting task: %s\n", url)
			result, err := svc.ProcessTask(context.Background(), url, opts, presenter)
			if err != nil {
				fmt.Printf("\nTask failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\nSuccess! Note saved to: %s\n", result.NotePath)
		},
	}

	// Define Flags
	rootCmd.Flags().BoolVarP(&audioOnly, "audio-only", "a", true, "Extract audio only (skip video download)")
	rootCmd.Flags().StringVar(&aiProvider, "ai-provider", "", "AI provider to use (ollama or openai)")
	rootCmd.Flags().StringVar(&model, "model", "", "LLM model name (e.g. qwen3:8b or gpt-4o)")
	rootCmd.Flags().StringVar(&translationMod, "translation-model", "", "Model used for translation (e.g. qwen3:0.6b)")
	rootCmd.Flags().StringVar(&targetLang, "target-lang", "", "Target language for analysis and translation")
	rootCmd.Flags().IntVar(&contextSize, "context-size", 0, "Context window size in tokens")
	rootCmd.Flags().StringVar(&vaultPath, "vault", "", "Override path to Obsidian Vault")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
