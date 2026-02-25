package app

import (
	"Varys/backend/analyzer"
	"Varys/backend/config"
	"Varys/backend/dependency"
	"Varys/backend/downloader"
	"Varys/backend/service"
	"Varys/backend/storage"
	"Varys/backend/transcriber"
	"Varys/backend/translation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
	coreService    *service.CoreService

	// Task Management
	taskMutex  sync.Mutex
	taskCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
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
	cfg := a.loadConfigSafe()

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

	// Initialize Core Service
	a.coreService = service.NewCoreService(a.depManager)

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

// WailsLogger implements service.EventLogger for the desktop app.
type WailsLogger struct {
	app       *App
	ctx       context.Context
	logBuffer []string
	hasError  bool
}

func (l *WailsLogger) Log(msg string) {
	if l.ctx.Err() != nil {
		return
	}
	l.logBuffer = append(l.logBuffer, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	runtime.EventsEmit(l.app.ctx, "task:log", msg)

	msgLower := strings.ToLower(msg)
	if strings.Contains(msgLower, "error") || strings.Contains(msgLower, "failed") {
		l.hasError = true
	}
}

func (l *WailsLogger) Progress(percentage float64) {
	if l.ctx.Err() == nil {
		runtime.EventsEmit(l.app.ctx, "task:progress", percentage)
	}
}

func (l *WailsLogger) AnalysisChunk(token string) {
	if l.ctx.Err() == nil {
		runtime.EventsEmit(l.app.ctx, "task:analysis", token)
	}
}

func (l *WailsLogger) Error(err error) {
	l.Log(fmt.Sprintf("Error: %v", err))
}

// SubmitTask starts the pipeline
func (a *App) SubmitTask(url string, audioOnly bool) (taskResult string, taskErr error) {
	runtime.LogInfo(a.ctx, fmt.Sprintf("Received task for URL: %s (AudioOnly: %v)", url, audioOnly))

	// Setup Cancellation Context
	a.taskMutex.Lock()
	if a.taskCancel != nil {
		a.taskCancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.taskCancel = cancel
	a.taskMutex.Unlock()

	defer func() {
		a.taskMutex.Lock()
		if a.taskCancel != nil {
			a.taskCancel()
			a.taskCancel = nil
		}
		a.taskMutex.Unlock()
	}()

	logger := &WailsLogger{app: a, ctx: ctx}

	// Dump logs on exit if error occurred (and not cancelled)
	defer func() {
		if ctx.Err() == context.Canceled {
			logger.Log("Task cancelled by user.")
			return
		}
		if taskErr != nil || logger.hasError {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			logContent := strings.Join(logger.logBuffer, "\n")
			if taskErr != nil {
				logContent += fmt.Sprintf("\n\n[FATAL ERROR] %v", taskErr)
			}

			home, _ := os.UserHomeDir()
			debugDir := filepath.Join(home, ".varys", "logs")
			os.MkdirAll(debugDir, 0755)

			filename := fmt.Sprintf("error_%s.log", timestamp)
			filePath := filepath.Join(debugDir, filename)
			os.WriteFile(filePath, []byte(logContent), 0644)

			latestPath := filepath.Join(debugDir, "error_latest.log")
			os.WriteFile(latestPath, []byte(logContent), 0644)

			logger.Log(fmt.Sprintf("Logs dumped to %s", filePath))
		}
	}()

	// Load latest config
	cfg := a.loadConfigSafe()

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

	if opts.ContextSize == 0 {
		opts.ContextSize = 8192
	}

	result, err := a.coreService.ProcessTask(ctx, url, opts, logger)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		taskErr = err
		return "", err
	}

	return fmt.Sprintf("Saved to: %s", result.NotePath), nil
}

// GetAppVersion returns the current application version
func (a *App) GetAppVersion() string {
	return "v0.4.1"
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

// LocateConfigFile opens the file explorer and selects the config file.
func (a *App) LocateConfigFile() error {
	path, err := a.GetConfigPath()
	if err != nil {
		return err
	}

	// For macOS (Darwin)
	return exec.Command("open", "-R", path).Run()
}

type DependencyStatus struct {
	YtDlp   bool `json:"yt_dlp"`
	Ffmpeg  bool `json:"ffmpeg"`
	Whisper bool `json:"whisper"`
	Ollama  bool `json:"ollama"`
}

type DiagnosticItem struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"` // ok | missing | misconfigured
	RequiredFor   []string `json:"required_for"`
	DetectedPath  string   `json:"detected_path"`
	FixSuggestion string   `json:"fix_suggestion"`
	FixCommands   []string `json:"fix_commands"`
	CanAutoFix    bool     `json:"can_auto_fix"`
	IsBlocker     bool     `json:"is_blocker"`
}

type StartupDiagnostics struct {
	GeneratedAt string           `json:"generated_at"`
	Provider    string           `json:"provider"`
	Blockers    []string         `json:"blockers"`
	Ready       bool             `json:"ready"`
	Items       []DiagnosticItem `json:"items"`
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

// GetStartupDiagnostics returns a unified environment + config readiness report.
func (a *App) GetStartupDiagnostics() StartupDiagnostics {
	diag := StartupDiagnostics{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Provider:    "ollama",
		Ready:       true,
		Items:       []DiagnosticItem{},
		Blockers:    []string{},
	}

	depStatus := a.CheckDependencies()
	cfg := &config.Config{}
	if a.cfgManager != nil {
		if loaded, err := a.cfgManager.Load(); err == nil && loaded != nil {
			cfg = loaded
		}
	}
	provider := cfg.AIProvider
	if provider == "" {
		provider = "ollama"
	}
	diag.Provider = provider

	addItem := func(item DiagnosticItem) {
		if item.IsBlocker {
			diag.Blockers = append(diag.Blockers, item.ID)
		}
		diag.Items = append(diag.Items, item)
	}

	addItem(buildBinaryItem(
		"yt-dlp",
		"yt-dlp",
		depStatus.YtDlp,
		[]string{"download"},
		true,
		"Install yt-dlp and ensure it is executable from PATH.",
		[]string{"brew install yt-dlp"},
	))

	addItem(buildBinaryItem(
		"ffmpeg",
		"ffmpeg",
		depStatus.Ffmpeg,
		[]string{"download", "transcribe"},
		true,
		"Install ffmpeg for audio conversion.",
		[]string{"brew install ffmpeg"},
	))

	addItem(buildBinaryItem(
		"whisper",
		"whisper.cpp",
		depStatus.Whisper,
		[]string{"transcribe"},
		true,
		"Install whisper.cpp and ensure the binary is available in PATH.",
		[]string{"brew install whisper-cpp"},
	))

	modelPath := strings.TrimSpace(cfg.ModelPath)
	modelOk := modelPath != ""
	if modelOk {
		if st, err := os.Stat(modelPath); err != nil || st.IsDir() {
			modelOk = false
		}
	}
	addItem(buildConfigItem(
		"model_path",
		"Whisper Model Path",
		modelOk,
		[]string{"transcribe"},
		true,
		modelPath,
		"Select an accessible Whisper model file (.bin) in Settings.",
		[]string{"In Settings, click Browse to select a Whisper model file."},
	))

	ollamaBlocker := provider == "ollama"
	ollamaRunning := checkOllamaRunning()
	addItem(buildOllamaItem(depStatus.Ollama, ollamaRunning, ollamaBlocker))
	addItem(buildOllamaModelsItem(ollamaRunning, ollamaBlocker))

	vaultPath := strings.TrimSpace(cfg.VaultPath)
	addItem(buildConfigItem(
		"vault_path",
		"Vault Path",
		vaultPath != "",
		[]string{"export"},
		true,
		vaultPath,
		"Select your Obsidian Vault directory in Settings.",
		[]string{"In Settings, click Browse to select your Obsidian Vault."},
	))

	openAIKey := strings.TrimSpace(cfg.OpenAIKey)
	openAIBlocker := provider == "openai"
	addItem(buildConfigItem(
		"openai_key",
		"OpenAI API Key",
		openAIKey != "",
		[]string{"analyze"},
		openAIBlocker,
		maskSecretForDisplay(openAIKey),
		"The current AI provider is OpenAI. Enter your OpenAI API key in Settings.",
		[]string{"Enter your OpenAI API key in Settings."},
	))

	diag.Ready = len(diag.Blockers) == 0
	return diag
}

func buildBinaryItem(id, name string, ok bool, requiredFor []string, blockerIfMissing bool, fixSuggestion string, fixCommands []string) DiagnosticItem {
	item := DiagnosticItem{
		ID:            id,
		Name:          name,
		RequiredFor:   requiredFor,
		FixSuggestion: fixSuggestion,
		FixCommands:   fixCommands,
		CanAutoFix:    false,
	}

	if ok {
		item.Status = "ok"
		item.DetectedPath = findBinaryPath(id)
		item.IsBlocker = false
	} else {
		item.Status = "missing"
		item.IsBlocker = blockerIfMissing
	}
	return item
}

func buildConfigItem(id, name string, ok bool, requiredFor []string, blockerIfBad bool, detectedPath, fixSuggestion string, fixCommands []string) DiagnosticItem {
	item := DiagnosticItem{
		ID:            id,
		Name:          name,
		RequiredFor:   requiredFor,
		FixSuggestion: fixSuggestion,
		FixCommands:   fixCommands,
		CanAutoFix:    false,
		DetectedPath:  detectedPath,
	}

	if ok {
		item.Status = "ok"
		item.IsBlocker = false
	} else {
		item.Status = "misconfigured"
		item.IsBlocker = blockerIfBad
	}

	return item
}

func buildOllamaItem(installed bool, running bool, blockerIfBad bool) DiagnosticItem {
	item := DiagnosticItem{
		ID:          "ollama",
		Name:        "Ollama",
		RequiredFor: []string{"analyze"},
		CanAutoFix:  false,
	}

	if !installed {
		item.Status = "missing"
		item.IsBlocker = blockerIfBad
		item.FixSuggestion = "The current AI provider is Ollama. Install Ollama first."
		item.FixCommands = []string{"brew install ollama"}
		return item
	}

	item.DetectedPath = findBinaryPath("ollama")
	if running {
		item.Status = "ok"
		item.IsBlocker = false
		item.FixSuggestion = "Ollama service is running normally."
		item.FixCommands = []string{}
		return item
	}

	item.Status = "misconfigured"
	item.IsBlocker = blockerIfBad
	item.CanAutoFix = true
	item.FixSuggestion = "Ollama is installed but not running. Start the service."
	item.FixCommands = []string{"ollama serve"}
	return item
}

func buildOllamaModelsItem(ollamaRunning bool, blockerIfBad bool) DiagnosticItem {
	modelsPath := getOllamaModelsPath()
	models, _ := getOllamaModelsFromAPI(ollamaRunning)
	hasModels := len(models) > 0 || hasAnyModelFiles(modelsPath)

	item := DiagnosticItem{
		ID:           "ollama_models",
		Name:         "Ollama Models",
		RequiredFor:  []string{"analyze", "translate"},
		DetectedPath: modelsPath,
		CanAutoFix:   false,
	}

	if hasModels {
		item.Status = "ok"
		item.IsBlocker = false
		item.FixSuggestion = "Available Ollama models were detected."
		return item
	}

	item.Status = "misconfigured"
	item.IsBlocker = blockerIfBad
	item.FixSuggestion = "No models were detected in the Ollama models path. Pull a model first."
	item.FixCommands = []string{
		"ollama pull qwen3:8b",
		"https://ollama.com/library",
	}
	return item
}

func findBinaryPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func checkOllamaRunning() bool {
	client := http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func getOllamaModelsPath() string {
	if p := strings.TrimSpace(os.Getenv("OLLAMA_MODELS")); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ollama", "models")
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func getOllamaModelsFromAPI(ollamaRunning bool) ([]string, error) {
	if !ollamaRunning {
		return nil, nil
	}
	client := http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var parsed ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		if strings.TrimSpace(m.Name) != "" {
			result = append(result, m.Name)
		}
	}
	return result, nil
}

func hasAnyModelFiles(modelsPath string) bool {
	if strings.TrimSpace(modelsPath) == "" {
		return false
	}
	if _, err := os.Stat(modelsPath); err != nil {
		return false
	}
	found := false
	_ = filepath.WalkDir(modelsPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func defaultConfig() *config.Config {
	return &config.Config{
		ContextSize:      8192,
		TranslationModel: "qwen3:0.6b",
		AIProvider:       "ollama",
		OpenAIModel:      "gpt-4o",
	}
}

// loadConfigSafe prevents startup/task crashes when config manager is unavailable
// or config.json is malformed (for example an empty file).
func (a *App) loadConfigSafe() *config.Config {
	if a.cfgManager == nil {
		if a.ctx != nil {
			runtime.LogWarning(a.ctx, "Config manager unavailable, falling back to default config.")
		}
		return defaultConfig()
	}

	cfg, err := a.cfgManager.Load()
	if err != nil || cfg == nil {
		if a.ctx != nil {
			runtime.LogWarningf(a.ctx, "Failed to load config, falling back to default config: %v", err)
		}
		return defaultConfig()
	}

	return cfg
}

// StartOllamaService tries to start `ollama serve` and waits briefly for readiness.
func (a *App) StartOllamaService() (string, error) {
	ollamaPath := findBinaryPath("ollama")
	if ollamaPath == "" {
		return "", fmt.Errorf("ollama binary not found in PATH")
	}

	if checkOllamaRunning() {
		return "ollama is already running", nil
	}

	cmd := exec.Command(ollamaPath, "serve")
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ollama serve: %w", err)
	}

	go func() {
		_ = cmd.Wait()
	}()

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		if checkOllamaRunning() {
			return "ollama started successfully", nil
		}
		time.Sleep(300 * time.Millisecond)
	}

	return "", fmt.Errorf("ollama process started but service is not ready yet")
}

// StopOllamaService tries to stop ollama server processes.
func (a *App) StopOllamaService() (string, error) {
	if !checkOllamaRunning() {
		return "ollama is already stopped", nil
	}

	// Stop all ollama processes started by `ollama serve`.
	cmd := exec.Command("pkill", "-x", "ollama")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to stop ollama process: %w", err)
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		if !checkOllamaRunning() {
			return "ollama stopped successfully", nil
		}
		time.Sleep(300 * time.Millisecond)
	}

	return "", fmt.Errorf("ollama stop command was sent but service is still responding")
}

// OpenOllamaModelLibrary opens Ollama model library in the default browser.
func (a *App) OpenOllamaModelLibrary() string {
	if a.ctx != nil {
		runtime.BrowserOpenURL(a.ctx, "https://ollama.com/library")
	}
	return "opened ollama model library"
}

// UpdateVaultPath updates vault path directly for setup wizard flow.
func (a *App) UpdateVaultPath(path string) error {
	if a.cfgManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	cfg := a.loadConfigSafe()
	cfg.VaultPath = strings.TrimSpace(path)
	return a.cfgManager.Save(cfg)
}

// UpdateModelPath updates whisper model path directly for setup wizard flow.
func (a *App) UpdateModelPath(path string) error {
	if a.cfgManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	cfg := a.loadConfigSafe()
	cfg.ModelPath = strings.TrimSpace(path)
	return a.cfgManager.Save(cfg)
}

// UpdateOpenAIKey updates OpenAI API key directly for setup wizard flow.
func (a *App) UpdateOpenAIKey(key string) error {
	if a.cfgManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	cfg := a.loadConfigSafe()
	cfg.OpenAIKey = strings.TrimSpace(key)
	return a.cfgManager.Save(cfg)
}

// ReadClipboardText returns text from system clipboard through Wails runtime.
func (a *App) ReadClipboardText() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}
	return runtime.ClipboardGetText(a.ctx)
}

func maskSecretForDisplay(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if len(text) <= 8 {
		return text
	}
	return text[:4] + strings.Repeat("*", len(text)-8) + text[len(text)-4:]
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
	if err := cfg.Validate(); err != nil {
		return err
	}
	return a.cfgManager.Save(&cfg)
}

// GetAIModels fetches available models from the selected AI provider
func (a *App) GetAIModels(providerType, apiKey string) ([]string, error) {
	an := analyzer.NewAnalyzer(providerType, apiKey, "")
	return an.ListModels(a.ctx)
}

// YtDlpUpdateInfo holds information about yt-dlp version status.
type YtDlpUpdateInfo struct {
	LocalVersion  string `json:"local_version"`
	LatestVersion string `json:"latest_version"`
	UpdateURL     string `json:"update_url"`
	HasUpdate     bool   `json:"has_update"`
}

// CheckYtDlpUpdate fetches latest version from GitHub and compares with local.
func (a *App) CheckYtDlpUpdate() (YtDlpUpdateInfo, error) {
	info := YtDlpUpdateInfo{
		UpdateURL: "https://github.com/yt-dlp/yt-dlp/releases/latest",
	}

	// 1. Get local version
	if a.depManager == nil {
		return info, fmt.Errorf("dependency manager not initialized")
	}

	ytPath := a.depManager.GetBinaryPath("yt-dlp")
	out, err := exec.Command(ytPath, "--version").Output()
	if err == nil {
		info.LocalVersion = strings.TrimSpace(string(out))
	}

	// 2. Fetch latest version from GitHub API (with timeout)
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(a.ctx, "GET", "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest", nil)
	if err != nil {
		return info, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return info, err
	}

	info.LatestVersion = release.TagName
	// Compare version strings (yt-dlp uses YYYY.MM.DD)
	if info.LocalVersion != "" && info.LatestVersion != "" {
		info.HasUpdate = info.LocalVersion != info.LatestVersion
	}

	return info, nil
}
