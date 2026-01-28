package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	VaultPath        string `json:"vault_path"`
	ModelPath        string `json:"model_path"` // Whisper Model Path
	LLMModel         string `json:"llm_model"`  // Ollama Model Name
	TranslationModel string `json:"translation_model"` // Ollama Model for Translation (Default: qwen3:0.6b)
	TargetLanguage   string `json:"target_language"` // Output language for analysis and translation
	ContextSize      int    `json:"context_size"`    // Context window size for Ollama (default: 8192)
}

type Manager struct {
	configPath string
}

func NewManager() (*Manager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir: %w", err)
	}

	appDir := filepath.Join(configDir, "Varys")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app config dir: %w", err)
	}

	return &Manager{
		configPath: filepath.Join(appDir, "config.json"),
	}, nil
}

func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if os.IsNotExist(err) {
		return &Config{ContextSize: 8192, TranslationModel: "qwen3:0.6b"}, nil // Return empty default config
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if missing
	if cfg.ContextSize == 0 {
		cfg.ContextSize = 8192
	}
	if cfg.TranslationModel == "" {
		cfg.TranslationModel = "qwen3:0.6b"
	}

	return &cfg, nil
}

func (m *Manager) Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (m *Manager) GetConfigPath() string {
	return m.configPath
}
