package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"Varys/backend/analyzer"
	"Varys/backend/config"
	"Varys/backend/dependency"
	"Varys/backend/downloader"
	"Varys/backend/storage"
	"Varys/backend/transcriber"

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
	
	// Analyzer
	llmModel := cfg.LLMModel
	if llmModel == "" { llmModel = "qwen2.5:7b" }
	a.analyzer = analyzer.NewAnalyzer(llmModel)

	runtime.LogInfo(a.ctx, "Backend initialized successfully.")
}

// SubmitTask starts the pipeline
func (a *App) SubmitTask(url string) (string, error) {
	runtime.LogInfo(a.ctx, "Received task for URL: "+url)

	tempDir, err := os.MkdirTemp("", "task_")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	// defer os.RemoveAll(tempDir) // Keep for debug

	// Load latest config
	cfg, _ := a.cfgManager.Load()

	// 0. Get Title
	runtime.EventsEmit(a.ctx, "task:log", "Fetching video title...")
	videoTitle, err := a.downloader.GetVideoTitle(url)
	if err != nil {
		videoTitle = "Task_" + time.Now().Format("20060102_150405")
		runtime.EventsEmit(a.ctx, "task:log", fmt.Sprintf("Warning: Failed to get title (%v), using fallback: %s", err, videoTitle))
	} else {
		runtime.EventsEmit(a.ctx, "task:log", fmt.Sprintf("Title found: %s", videoTitle))
	}

	// 1. Download
	runtime.EventsEmit(a.ctx, "task:log", fmt.Sprintf("Downloading audio from %s...", url))
	audioPath, err := a.downloader.DownloadAudio(url, tempDir, func(msg string) {
		runtime.EventsEmit(a.ctx, "task:log", "[DL] "+msg)
	})
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	runtime.EventsEmit(a.ctx, "task:log", fmt.Sprintf("Download complete: %s", audioPath))

	// 2. Transcribe
	runtime.EventsEmit(a.ctx, "task:log", "Transcribing audio (this may take a while)...")
	// Pass model from config dynamically
	transcript, err := a.transcriber.Transcribe(audioPath, cfg.ModelPath, func(msg string) {
		runtime.EventsEmit(a.ctx, "task:log", "[Whisper] "+msg)
	})
	if err != nil {
		runtime.LogErrorf(a.ctx, "Transcription failed: %v", err)
		runtime.EventsEmit(a.ctx, "task:log", "Transcription failed (check config/deps). Skipping analysis.")
		transcript = "Transcription failed."
	} else {
		runtime.EventsEmit(a.ctx, "task:log", "Transcription complete.")
	}

	// 3. Analyze
	// Reload analyzer with latest model if needed?
	// For simplicity, we create a new analyzer or update it.
	// Since Analyzer struct holds modelName, let's just make a new one or update it.
	// Analyzer is lightweight.
	llmModel := cfg.LLMModel
	if llmModel == "" { llmModel = "qwen2.5:7b" }
	localAnalyzer := analyzer.NewAnalyzer(llmModel)

	var summary string
	if transcript != "Transcription failed." {
		runtime.EventsEmit(a.ctx, "task:log", "Analyzing text with AI...")
		analysis, err := localAnalyzer.Analyze(transcript)
		if err != nil {
			runtime.LogErrorf(a.ctx, "Analysis failed: %v", err)
			runtime.EventsEmit(a.ctx, "task:log", "Analysis failed (is Ollama running?).")
			summary = "Analysis failed."
		} else {
			summary = analysis
			runtime.EventsEmit(a.ctx, "task:log", "Analysis complete.")
		}
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
	
	finalAudio, err := localStorage.MoveAudio(audioPath, safeTitle)
	if err != nil {
		return "", fmt.Errorf("failed to move audio: %w", err)
	}

	noteData := storage.NoteData{
		Title:        safeTitle,
		URL:          url,
		Language:     "en",
		Summary:      summary,
		OriginalText: transcript,
		AudioFile:    finalAudio,
		AssetsFolder: "assets",
		CreatedTime:  time.Now().Format("2006-01-02 15:04"),
	}

	notePath, err := localStorage.SaveNote(noteData)
	if err != nil {
		return "", fmt.Errorf("failed to save note: %w", err)
	}

	return fmt.Sprintf("Saved to: %s", notePath), nil
}

// GetConfig returns current config
func (a *App) GetConfig() (*config.Config, error) {
	if a.cfgManager == nil {
		return nil, fmt.Errorf("config manager not initialized")
	}
	return a.cfgManager.Load()
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