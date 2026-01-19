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
