//go:build integration
package app

import (
	"Varys/backend/config"
	"os"
	"path/filepath"
	"testing"
)

func TestGetStartupDiagnostics_OpenAI_OllamaNotBlocker(t *testing.T) {
	app := buildTestAppWithConfig(t, config.Config{
		AIProvider: "openai",
		OpenAIKey:  "sk-test",
		VaultPath:  t.TempDir(),
		ModelPath:  createTempModel(t),
	})

	diag := app.GetStartupDiagnostics()
	ollama := findDiagnosticItem(t, diag, "ollama")
	openaiKey := findDiagnosticItem(t, diag, "openai_key")

	if ollama.IsBlocker {
		t.Fatalf("expected ollama to be non-blocker for openai provider")
	}
	if openaiKey.IsBlocker {
		t.Fatalf("expected openai_key to be non-blocker when key is configured")
	}
	if openaiKey.Status != "ok" {
		t.Fatalf("expected openai_key status=ok, got=%s", openaiKey.Status)
	}
}

func TestGetStartupDiagnostics_OpenAI_MissingKeyIsBlocker(t *testing.T) {
	app := buildTestAppWithConfig(t, config.Config{
		AIProvider: "openai",
		VaultPath:  t.TempDir(),
		ModelPath:  createTempModel(t),
	})

	diag := app.GetStartupDiagnostics()
	openaiKey := findDiagnosticItem(t, diag, "openai_key")

	if !openaiKey.IsBlocker {
		t.Fatalf("expected openai_key to be blocker when provider=openai and key missing")
	}
	if openaiKey.Status != "misconfigured" {
		t.Fatalf("expected openai_key status=misconfigured, got=%s", openaiKey.Status)
	}
}

func TestGetStartupDiagnostics_Ollama_MissingIsBlocker(t *testing.T) {
	app := buildTestAppWithConfig(t, config.Config{
		AIProvider: "ollama",
		VaultPath:  t.TempDir(),
		ModelPath:  createTempModel(t),
	})

	diag := app.GetStartupDiagnostics()
	ollama := findDiagnosticItem(t, diag, "ollama")

	if !ollama.IsBlocker {
		t.Fatalf("expected ollama to be blocker when provider=ollama and binary missing")
	}
	if ollama.Status != "missing" {
		t.Fatalf("expected ollama status=missing, got=%s", ollama.Status)
	}
}

func TestGetStartupDiagnostics_VaultAndModelMissingAreBlockers(t *testing.T) {
	app := buildTestAppWithConfig(t, config.Config{
		AIProvider: "openai",
		OpenAIKey:  "sk-test",
	})

	diag := app.GetStartupDiagnostics()
	vault := findDiagnosticItem(t, diag, "vault_path")
	model := findDiagnosticItem(t, diag, "model_path")

	if !vault.IsBlocker || vault.Status != "misconfigured" {
		t.Fatalf("expected vault_path to be misconfigured blocker, got blocker=%v status=%s", vault.IsBlocker, vault.Status)
	}
	if !model.IsBlocker || model.Status != "misconfigured" {
		t.Fatalf("expected model_path to be misconfigured blocker, got blocker=%v status=%s", model.IsBlocker, model.Status)
	}
}

func TestLoadConfigSafe_DefaultWhenConfigInvalid(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	mgr, err := config.NewManager()
	if err != nil {
		t.Fatalf("failed to init config manager: %v", err)
	}

	// Simulate corrupted/empty config.json
	if err := os.WriteFile(mgr.GetConfigPath(), []byte(""), 0644); err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	app := &App{cfgManager: mgr}
	cfg := app.loadConfigSafe()

	if cfg == nil {
		t.Fatal("expected fallback config, got nil")
	}
	if cfg.ContextSize != 8192 {
		t.Fatalf("expected default context size 8192, got %d", cfg.ContextSize)
	}
	if cfg.TranslationModel != "qwen3:0.6b" {
		t.Fatalf("expected default translation model qwen3:0.6b, got %s", cfg.TranslationModel)
	}
	if cfg.AIProvider != "ollama" {
		t.Fatalf("expected default provider ollama, got %s", cfg.AIProvider)
	}
}

func TestGetStartupDiagnostics_OllamaModelsMissingBlocksWhenProviderOllama(t *testing.T) {
	if checkOllamaRunning() {
		t.Skip("ollama is running with real models; skip deterministic empty-models test")
	}

	modelsDir := t.TempDir()
	t.Setenv("OLLAMA_MODELS", modelsDir)

	app := buildTestAppWithConfig(t, config.Config{
		AIProvider: "ollama",
		VaultPath:  t.TempDir(),
		ModelPath:  createTempModel(t),
	})

	diag := app.GetStartupDiagnostics()
	item := findDiagnosticItem(t, diag, "ollama_models")

	if item.Status != "misconfigured" {
		t.Fatalf("expected ollama_models status=misconfigured, got=%s", item.Status)
	}
	if !item.IsBlocker {
		t.Fatalf("expected ollama_models to be blocker for ollama provider")
	}
}

func buildTestAppWithConfig(t *testing.T, cfg config.Config) *App {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	mgr, err := config.NewManager()
	if err != nil {
		t.Fatalf("failed to init config manager: %v", err)
	}
	if err := mgr.Save(&cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	return &App{cfgManager: mgr}
}

func createTempModel(t *testing.T) string {
	t.Helper()
	modelPath := filepath.Join(t.TempDir(), "ggml-base.bin")
	if err := os.WriteFile(modelPath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to create temp model: %v", err)
	}
	return modelPath
}

func findDiagnosticItem(t *testing.T, diag StartupDiagnostics, id string) DiagnosticItem {
	t.Helper()
	for _, item := range diag.Items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("diagnostic item not found: %s", id)
	return DiagnosticItem{}
}
