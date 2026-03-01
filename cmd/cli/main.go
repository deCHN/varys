package main

import (
	"Varys/backend/config"
	"Varys/backend/dependency"
	"Varys/backend/search"
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

var (
	videoOnly      bool
	aiProvider     string
	model          string
	translationMod string
	targetLang     string
	contextSize    int
	vaultPath      string
	searchLimit    int
	searchProvider string
)

func runTask(url string, cmd *cobra.Command) {
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
		AudioOnly:      !videoOnly,
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
	if v, err := cmd.Flags().GetBool("video"); err == nil && cmd.Flags().Changed("video") {
		opts.AudioOnly = !v
	} else if cmd.Flags().Changed("video") {
		opts.AudioOnly = !videoOnly
	}
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
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "varys-cli [URL or local path]",
		Short: "Varys CLI - Transcribe and analyze audio/video content",
		Long:  `Varys is a local-first desktop and CLI application designed to automate the capture, transcription, and analysis of video and audio content.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			runTask(args[0], cmd)
		},
	}

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search and ingest content from various platforms",
		Long: `Search for video, audio, or web content using various providers (like yt-dlp or Tavily).
An interactive TUI will open to show the results:
- [Space]: Toggle selection for multiple items (Marked with [x]).
- [Enter]: Confirm and start processing all selected items.
- [q/Ctrl+C]: Quit.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]

			// Load Config for API keys
			cm, _ := config.NewManager()
			cfg, _ := cm.Load()

			sm := search.NewSearchManager(os.Getenv("TAVILY_API_KEY"))
			if cfg.OpenAIKey != "" && os.Getenv("TAVILY_API_KEY") == "" {
				// Use key from config if env is not set
				// sm = search.NewSearchManager(cfg.OpenAIKey) // Wait, Tavily key is separate
			}

			p, err := sm.GetProvider(searchProvider)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			opts := search.SearchOptions{
				Limit: searchLimit,
				Type:  search.ContentTypeAll,
			}

			choices, err := RunSearchTUI(query, p, opts)
			if err != nil {
				fmt.Printf("Search failed: %v\n", err)
				os.Exit(1)
			}

			for _, choice := range choices {
				fmt.Printf("\nSelected: %s (%s)\n", choice.Title, choice.URL)
				runTask(choice.URL, cmd)
			}
		},
	}

	// Root Flags
	rootCmd.PersistentFlags().BoolVarP(&videoOnly, "video", "v", false, "Download full video instead of audio only")
	rootCmd.PersistentFlags().StringVarP(&aiProvider, "ai-provider", "p", "", "AI provider to use (ollama or Cloud LLMs)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "", "LLM model name")
	rootCmd.PersistentFlags().StringVar(&translationMod, "translation-model", "", "Model used for translation")
	rootCmd.PersistentFlags().StringVar(&targetLang, "target-lang", "", "Target language for analysis and translation")
	rootCmd.PersistentFlags().IntVar(&contextSize, "context-size", 0, "Context window size in tokens")
	rootCmd.PersistentFlags().StringVar(&vaultPath, "vault", "", "Override path to Obsidian Vault")

	// Search Flags
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 5, "Number of search results to fetch")
	searchCmd.Flags().StringVarP(&searchProvider, "provider", "s", "yt-dlp", "Search provider (yt-dlp, tavily)")

	rootCmd.AddCommand(searchCmd)

	// Completion
	rootCmd.RegisterFlagCompletionFunc("ai-provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"ollama", "openai"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.RegisterFlagCompletionFunc("target-lang", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"Simplified Chinese", "Traditional Chinese", "English", "Japanese", "French", "German"}, cobra.ShellCompDirectiveNoFileComp
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
