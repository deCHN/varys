package config

import (
	"os"
	"testing"
)

func TestConfigSaveLoad(t *testing.T) {
	// Use a temp file for testing by overriding NewManager logic or just manually setting path
	// Since NewManager uses UserHomeDir, let's just test Load/Save with a custom Manager

	f, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	mgr := &Manager{configPath: f.Name()}

	// Test Save
	cfg := &Config{
		VaultPath: "/tmp/vault",
		ModelPath: "/tmp/models",
	}
	if err := mgr.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Load
	loaded, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.VaultPath != cfg.VaultPath {
		t.Errorf("Expected VaultPath %s, got %s", cfg.VaultPath, loaded.VaultPath)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Ollama Config",
			config: Config{
				VaultPath:  "/path/to/vault",
				AIProvider: "ollama",
			},
			wantErr: false,
		},
		{
			name: "Valid OpenAI Config",
			config: Config{
				VaultPath:  "/path/to/vault",
				AIProvider: "openai",
				OpenAIKey:  "sk-123",
			},
			wantErr: false,
		},
		{
			name: "Missing Vault Path",
			config: Config{
				AIProvider: "ollama",
			},
			wantErr: true,
		},
		{
			name: "Missing OpenAI Key",
			config: Config{
				VaultPath:  "/path/to/vault",
				AIProvider: "openai",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
