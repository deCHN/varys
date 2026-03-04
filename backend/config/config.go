package config

import (
	"Varys/backend/secret"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	VaultPath        string `json:"vault_path"`
	ModelPath        string `json:"model_path"`        // Whisper Model Path
	LLMModel         string `json:"llm_model"`         // Ollama Model Name
	TranslationModel string `json:"translation_model"` // Ollama Model for Translation (Default: qwen3:0.6b)
	TargetLanguage   string `json:"target_language"`   // Output language for analysis and translation
	ContextSize      int    `json:"context_size"`      // Context window size for Ollama (default: 8192)
	CustomPrompt     string `json:"custom_prompt"`     // Custom user prompt for analysis
	AIProvider       string `json:"ai_provider"`       // "ollama" or "openai"
	OpenAIModel      string `json:"openai_model"`      // e.g. "gpt-4o"
	OpenAIKey        string `json:"openai_key,omitempty"` // Stored in Keyring, passed via Wails
}

type Manager struct {
	configPath string
	appDir     string
}

var (
	// userHomeDir is a var so it can be mocked in tests
	userHomeDir = os.UserHomeDir
)

// GetConfigDir returns the directory for config.json
// Priority:
// 1. Environment variable: VARYS_CONFIG_DIR
// 2. XDG/CLI habit: ~/.config/varys (if config.json exists there)
// 3. System default fallback: os.UserConfigDir()
func GetConfigDir() (string, error) {
	// 1. Check environment variable
	if envDir := os.Getenv("VARYS_CONFIG_DIR"); envDir != "" {
		return envDir, nil
	}

	home, _ := userHomeDir()
	xdgDir := filepath.Join(home, ".config", "Varys")
	xdgConfig := filepath.Join(xdgDir, "config.json")

	// 2. Check XDG/CLI habit (if config.json exists)
	if _, err := os.Stat(xdgConfig); err == nil {
		return xdgDir, nil
	}

	// 3. System default fallback
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	return filepath.Join(configDir, "Varys"), nil
}

// GetLogDir returns the system standard directory for logs.
// On macOS: ~/Library/Logs/Varys
func GetLogDir() (string, error) {
	home, _ := userHomeDir()
	// macOS Standard Log Path
	logDir := filepath.Join(home, "Library", "Logs", "Varys")
	
	// Fallback for non-macOS or if Library/Logs doesn't exist
	if _, err := os.Stat(filepath.Join(home, "Library", "Logs")); err != nil {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		logDir = filepath.Join(cacheDir, "Varys", "logs")
	}
	
	return logDir, nil
}

func NewManager() (*Manager, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	logDir, err := GetLogDir()
	if err != nil {
		return nil, err
	}

	// Ensure the app directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	// Ensure the log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Manager{
		configPath: filepath.Join(configDir, "config.json"),
		appDir:     configDir,
	}, nil
}

func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if os.IsNotExist(err) {
		return &Config{ContextSize: 8192, TranslationModel: "qwen3:0.6b"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 1. Parse base config
	// Note: OpenAIKey will be empty here because of json:"-"
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 2. Fetch secrets from Keyring
	key, err := secret.GetSecret(secret.KeyAccountOpenAI)
	if err == nil && key != "" {
		cfg.OpenAIKey = key
	}

	// 3. Migration: Check if config.json still contains plain-text keys
	// We do this by re-parsing into a map to see if the field exists in JSON
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if oldKey, ok := raw["openai_key"].(string); ok && oldKey != "" {
		// Found legacy key, move to keyring
		if err := secret.SetSecret(secret.KeyAccountOpenAI, oldKey); err == nil {
			cfg.OpenAIKey = oldKey
			// Trigger a save to clean up the JSON file
			m.Save(&cfg)
		}
	}

	// Set defaults if missing
	if cfg.ContextSize == 0 {
		cfg.ContextSize = 8192
	}
	if cfg.TranslationModel == "" {
		cfg.TranslationModel = "qwen3:0.6b"
	}
	if cfg.AIProvider == "" {
		cfg.AIProvider = "ollama"
	}
	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = "gpt-4o"
	}
	if cfg.TargetLanguage == "" {
		cfg.TargetLanguage = "English"
	}

	return &cfg, nil
}

func (m *Manager) Save(cfg *Config) error {
	// 1. Save sensitive data to Keyring
	if cfg.OpenAIKey != "" {
		if err := secret.SetSecret(secret.KeyAccountOpenAI, cfg.OpenAIKey); err != nil {
			return fmt.Errorf("failed to save secret to keyring: %w", err)
		}
	}

	// 2. Prepare a copy for file storage (without sensitive keys)
	fileCfg := *cfg
	fileCfg.OpenAIKey = "" // Clear before saving to disk

	data, err := json.MarshalIndent(fileCfg, "", "  ")
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

func (c *Config) Validate() error {
	if c.VaultPath == "" {
		return fmt.Errorf("obsidian vault path is required")
	}
	if c.AIProvider == "openai" && c.OpenAIKey == "" {
		return fmt.Errorf("openai api key is required when openai provider is selected")
	}
	return nil
}
