package main

import (
	"Varys/backend/analyzer"
	"Varys/backend/config"
	"Varys/backend/dependency"
	"Varys/backend/downloader"
	"Varys/backend/storage"
	"Varys/backend/transcriber"
	"Varys/backend/translation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	depManager     *dependency.Manager
	cfgManager     *config.Manager
	downloader     *downloader.Downloader
	storageManager *storage.Manager
	transcriber    *transcriber.Transcriber
	analyzer       *analyzer.Analyzer
	translator     *translation.Translator

	// Task Management
	taskMutex  sync.Mutex
	taskCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Dependencies
	dm, err := dependency.NewManager()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Error initializing dependency manager: %v", err)
		return
	}
	a.depManager = dm
	// We only ensure yt-dlp/ffmpeg now. Whisper/Llama are system deps.
	if err := a.depManager.EnsureBinaries(); err != nil {
		runtime.LogErrorf(a.ctx, "Error ensuring binaries: %v", err)
	}

	// 2. Config
	cm, err := config.NewManager()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Error initializing config manager: %v", err)
	}
	a.cfgManager = cm

	// 3. Init Modules
	cfg, _ := a.cfgManager.Load()

	// Storage
	vaultPath := cfg.VaultPath
	if vaultPath == "" {
		home, _ := os.UserHomeDir()
		vaultPath = home
	}
	a.storageManager = storage.NewManager(vaultPath)

	// Downloader
	a.downloader = downloader.NewDownloader(a.depManager)

	// Transcriber (Model passed dynamically now)
	a.transcriber = transcriber.NewTranscriber(a.depManager)

	// Analyzer & Translator
	llmModel := cfg.LLMModel
	translationModel := cfg.TranslationModel
	if llmModel == "" {
		llmModel = "qwen3:8b"
	}
	if translationModel == "" {
		translationModel = "qwen3:0.6b"
	}

	aiProvider := cfg.AIProvider
	analyzerModel := llmModel
	if aiProvider == "openai" {
		analyzerModel = cfg.OpenAIModel
		if analyzerModel == "" {
			analyzerModel = "gpt-4o"
		}
	}

	a.analyzer = analyzer.NewAnalyzer(aiProvider, cfg.OpenAIKey, analyzerModel)
	a.translator = translation.NewTranslator(translationModel)

	runtime.LogInfo(a.ctx, "Backend initialized successfully.")
}

// CancelTask cancels the currently running task
func (a *App) CancelTask() {
	a.taskMutex.Lock()
	defer a.taskMutex.Unlock()
	if a.taskCancel != nil {
		a.taskCancel()
		a.taskCancel = nil
		runtime.LogInfo(a.ctx, "Task cancellation requested.")
	}
}

// SubmitTask starts the pipeline
func (a *App) SubmitTask(url string, audioOnly bool) (taskResult string, taskErr error) {
	runtime.LogInfo(a.ctx, fmt.Sprintf("Received task for URL: %s (AudioOnly: %v)", url, audioOnly))

	// Setup Cancellation Context
	a.taskMutex.Lock()
	// Cancel previous if any
	if a.taskCancel != nil {
		a.taskCancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.taskCancel = cancel
	a.taskMutex.Unlock()

	// Ensure cleanup
	defer func() {
		a.taskMutex.Lock()
		if a.taskCancel != nil {
			a.taskCancel() // Ensure context is cancelled
			a.taskCancel = nil
		}
		a.taskMutex.Unlock()
	}()

	var logBuffer []string
	hasErrorLog := false

	// logFunc captures logs and emits them
	logFunc := func(msg string) {
		// Check cancellation before logging
		if ctx.Err() != nil {
			return
		}

		logBuffer = append(logBuffer, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
		runtime.EventsEmit(a.ctx, "task:log", msg)

		msgLower := strings.ToLower(msg)
		if strings.Contains(msgLower, "error") || strings.Contains(msgLower, "failed") {
			hasErrorLog = true
		}
	}

	// Dump logs on exit if error occurred (and not cancelled)
	defer func() {
		if ctx.Err() == context.Canceled {
			logFunc("Task cancelled by user.")
			return
		}
		if taskErr != nil || hasErrorLog {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			logContent := strings.Join(logBuffer, "\n")
			if taskErr != nil {
				logContent += fmt.Sprintf("\n\n[FATAL ERROR] %v", taskErr)
			}

			// Save to ~/.varys/logs
			home, _ := os.UserHomeDir()
			debugDir := filepath.Join(home, ".varys", "logs")
			os.MkdirAll(debugDir, 0755)

			filename := fmt.Sprintf("error_%s.log", timestamp)
			filePath := filepath.Join(debugDir, filename)
			os.WriteFile(filePath, []byte(logContent), 0644)

			// Copy to error_latest.log
			latestPath := filepath.Join(debugDir, "error_latest.log")
			os.WriteFile(latestPath, []byte(logContent), 0644)

			logFunc(fmt.Sprintf("Logs dumped to %s", filePath))
		}
	}()

	tempDir, err := os.MkdirTemp("", "task_")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp files

	// Check Cancel
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Load latest config
	cfg, _ := a.cfgManager.Load()

	// 0. Get Title
	logFunc("Fetching video title...")
	videoTitle, err := a.downloader.GetVideoTitle(url)
	if err != nil {
		videoTitle = "Task_" + time.Now().Format("20060102_150405")
		logFunc(fmt.Sprintf("Warning: Failed to get title (%v), using fallback: %s", err, videoTitle))
	} else {
		logFunc(fmt.Sprintf("Title found: %s", videoTitle))
	}

	// Fetch Description
	logFunc("Fetching video description...")
	videoDescription, err := a.downloader.GetVideoDescription(url)
	if err != nil {
		logFunc(fmt.Sprintf("Warning: Failed to get description: %v", err))
		videoDescription = ""
	}

	// Check Cancel
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// 1. Download
	logFunc(fmt.Sprintf("Downloading media from %s...", url))
	mediaPath, err := a.downloader.DownloadMedia(url, tempDir, audioOnly, func(msg string) {
		if ctx.Err() == nil {
			logFunc("[DL] " + msg)
		}
	})
	if err != nil {
		// If context cancelled, err might be from downloader or our check
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("download failed: %w", err)
	}
	logFunc(fmt.Sprintf("Download complete: %s", mediaPath))

	// Check Cancel
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// 2. Transcribe
	logFunc("Transcribing audio (this may take a while)...")
	var sourceLang string
	// Pass model from config dynamically
	transcript, sourceLang, err := a.transcriber.Transcribe(mediaPath, cfg.ModelPath, func(msg string) {
		if ctx.Err() == nil {
			logFunc("[Whisper] " + msg)
		}
	})
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		runtime.LogErrorf(a.ctx, "Transcription failed: %v", err)
		logFunc("Transcription failed (check config/deps). Skipping analysis.")
		transcript = "Transcription failed."
	} else {
		logFunc(fmt.Sprintf("Transcription complete (Detected Language: %s).", sourceLang))
	}

	// Check Cancel
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// 3. Analyze & Translate
	llmModel := cfg.LLMModel
	translationModel := cfg.TranslationModel
	if llmModel == "" {
		llmModel = "qwen3:8b"
	}
	if translationModel == "" {
		translationModel = "qwen3:0.6b"
	}

	aiProvider := cfg.AIProvider
	analyzerModel := llmModel
	if aiProvider == "openai" {
		analyzerModel = cfg.OpenAIModel
		if analyzerModel == "" {
			analyzerModel = "gpt-4o"
		}
	}

	localAnalyzer := analyzer.NewAnalyzer(aiProvider, cfg.OpenAIKey, analyzerModel)
	localTranslator := translation.NewTranslator(translationModel)

	targetLang := cfg.TargetLanguage
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}

	contextSize := cfg.ContextSize
	if contextSize == 0 {
		contextSize = 8192
	}

	var summary string
	var analysis *analyzer.AnalysisResult
	var translationPairs []translation.TranslationPair

	if transcript != "Transcription failed." {
		// Smart Translation Logic
		shouldTranslate := true

		// 1. Check if source matches target (Basic mapping)
		isChineseSource := sourceLang == "zh"
		isChineseTarget := strings.Contains(targetLang, "Chinese")

		isEnglishSource := sourceLang == "en"
		isEnglishTarget := strings.Contains(targetLang, "English")

		if (isChineseSource && isChineseTarget) || (isEnglishSource && isEnglishTarget) {
			shouldTranslate = false
			logFunc("Source language matches target. Skipping translation.")
		}

		if shouldTranslate {
			// Check Cancel
			if ctx.Err() != nil {
				return "", ctx.Err()
			}

			// A. Translate
			logFunc(fmt.Sprintf("Translating text to %s (structured)...", targetLang))
			var err error
			translationPairs, err = localTranslator.Translate(transcript, targetLang, contextSize, func(current, total int) {
				if ctx.Err() == nil {
					percent := float64(current+1) / float64(total) * 100
					runtime.EventsEmit(a.ctx, "task:progress", percent)
				}
			})
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			if err != nil {
				runtime.LogErrorf(a.ctx, "Translation failed: %v", err)
				logFunc(fmt.Sprintf("Translation failed: %v", err))
			} else {
				runtime.EventsEmit(a.ctx, "task:progress", 100.0) // Ensure it finishes
				logFunc("Translation complete.")
			}
		}

		// Check Cancel
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// B. Analyze
		logFunc("Analyzing content...")
		analysis, err = localAnalyzer.Analyze(ctx, transcript, cfg.CustomPrompt, targetLang, contextSize, func(token string) {
			if ctx.Err() == nil {
				runtime.EventsEmit(a.ctx, "task:analysis", token)
			}
		})
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err != nil {
			runtime.LogErrorf(a.ctx, "Analysis failed: %v", err)
			logFunc(fmt.Sprintf("Analysis failed: %v", err))
			summary = "Analysis failed."
			analysis = &analyzer.AnalysisResult{}
		} else {
			summary = analysis.Summary
			logFunc("Analysis complete.")
		}
	}

	// Check Cancel
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// 4. Save
	// Reload storage manager too? VaultPath might have changed.
	vaultPath := cfg.VaultPath
	if vaultPath == "" {
		home, _ := os.UserHomeDir()
		vaultPath = home
	}
	localStorage := storage.NewManager(vaultPath)

	safeTitle := localStorage.SanitizeFilename(videoTitle)

	finalMedia, err := localStorage.MoveMedia(mediaPath, safeTitle)
	if err != nil {
		return "", fmt.Errorf("failed to move media: %w", err)
	}

	noteData := storage.NoteData{
		Title:            safeTitle,
		URL:              url,
		Language:         targetLang, // Use target language for note context
		Description:      videoDescription,
		Summary:          summary,
		KeyPoints:        analysis.KeyPoints,
		Tags:             analysis.Tags,
		Assessment:       analysis.Assessment,
		OriginalText:     transcript,
		TranslationPairs: translationPairs,
		AudioFile:        finalMedia,
		AssetsFolder:     "assets",
		CreatedTime:      time.Now().Format("2006-01-02 15:04"),
	}

	notePath, err := localStorage.SaveNote(noteData)
	if err != nil {
		return "", fmt.Errorf("failed to save note: %w", err)
	}

	return fmt.Sprintf("Saved to: %s", notePath), nil
}

// GetAppVersion returns the current application version
func (a *App) GetAppVersion() string {
	return "v0.3.8"
}

// GetConfig returns current config
func (a *App) GetConfig() (*config.Config, error) {
	if a.cfgManager == nil {
		return nil, fmt.Errorf("config manager not initialized")
	}
	return a.cfgManager.Load()
}

// GetConfigPath returns the absolute path to the config file
func (a *App) GetConfigPath() (string, error) {
	if a.cfgManager == nil {
		return "", fmt.Errorf("config manager not initialized")
	}
	return a.cfgManager.GetConfigPath(), nil
}

type DependencyStatus struct {
	YtDlp   bool `json:"yt_dlp"`
	Ffmpeg  bool `json:"ffmpeg"`
	Whisper bool `json:"whisper"`
	Ollama  bool `json:"ollama"`
}

// CheckDependencies checks availability of external tools
func (a *App) CheckDependencies() DependencyStatus {
	status := DependencyStatus{}

	if a.depManager == nil {
		return status
	}

	// 1. Embedded/Sidecar
	ytPath := a.depManager.GetBinaryPath("yt-dlp")
	if _, err := os.Stat(ytPath); err == nil {
		status.YtDlp = true
	}

	ffmpegPath := a.depManager.GetBinaryPath("ffmpeg")
	if _, err := os.Stat(ffmpegPath); err == nil {
		status.Ffmpeg = true
	} else {
		// Check system ffmpeg
		if _, found := a.depManager.CheckSystemDependency("ffmpeg"); found {
			status.Ffmpeg = true
		}
	}

	// 2. System AI
	candidates := []string{"whisper-cpp", "whisper-cli", "whisper-main", "whisper", "main"}
	for _, name := range candidates {
		if _, found := a.depManager.CheckSystemDependency(name); found {
			status.Whisper = true
			break
		}
	}

	// 3. Ollama
	if _, found := a.depManager.CheckSystemDependency("ollama"); found {
		status.Ollama = true
	}

	return status
}

// SelectVaultPath opens a directory dialog
func (a *App) SelectVaultPath() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Obsidian Vault Folder",
	})
}

// SelectModelPath opens a file dialog for .bin
func (a *App) SelectModelPath() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Whisper Model",
		Filters: []runtime.FileFilter{
			{DisplayName: "Whisper Models (*.bin)", Pattern: "*.bin"},
		},
	})
}

// UpdateConfig saves the configuration
func (a *App) UpdateConfig(cfg config.Config) error {
	if a.cfgManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.cfgManager.Save(&cfg)
}

type OllamaModel struct {
	Name string `json:"name"`
}
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// GetOllamaModels fetches available models from local Ollama instance
func (a *App) GetOllamaModels() ([]string, error) {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ollama: %w", err)
	}
	defer resp.Body.Close()

	var data OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range data.Models {
		names = append(names, m.Name)
	}
	return names, nil
}
