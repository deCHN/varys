package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	
	// 1. Test Environment Variable Priority
	t.Run("EnvVarPriority", func(t *testing.T) {
		expected := "/tmp/varys_custom_config"
		os.Setenv("VARYS_CONFIG_DIR", expected)
		defer os.Unsetenv("VARYS_CONFIG_DIR")

		got, err := GetConfigDir()
		if err != nil {
			t.Errorf("GetConfigDir failed: %v", err)
		}
		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}
	})

	// 2. Test XDG/CLI habit (if config.json exists)
	t.Run("XDGHabitPriority", func(t *testing.T) {
		os.Unsetenv("VARYS_CONFIG_DIR")
		
		// Create a mock XDG config directory and file
		xdgDir := filepath.Join(home, ".config", "varys")
		os.MkdirAll(xdgDir, 0755)
		configPath := filepath.Join(xdgDir, "config.json")
		os.WriteFile(configPath, []byte("{}"), 0644)
		defer os.Remove(configPath)

		got, err := GetConfigDir()
		if err != nil {
			t.Errorf("GetConfigDir failed: %v", err)
		}
		if got != xdgDir {
			t.Errorf("Expected %s, got %s", xdgDir, got)
		}
	})

	// 3. Test System Default Fallback
	t.Run("SystemDefaultFallback", func(t *testing.T) {
		os.Unsetenv("VARYS_CONFIG_DIR")
		
		// Ensure XDG config file doesn't exist for this test
		xdgConfig := filepath.Join(home, ".config", "varys", "config.json")
		os.Remove(xdgConfig)

		got, err := GetConfigDir()
		if err != nil {
			t.Errorf("GetConfigDir failed: %v", err)
		}
		
		// On macOS it should be ~/Library/Application Support/Varys
		if !strings.Contains(got, "Application Support") && !strings.Contains(got, "Varys") {
			t.Errorf("Expected system default path, got %s", got)
		}
	})
}

func TestGetLogDir(t *testing.T) {
	got, err := GetLogDir()
	if err != nil {
		t.Errorf("GetLogDir failed: %v", err)
	}

	// On macOS, it should be in ~/Library/Logs/Varys
	if !strings.Contains(got, "Library/Logs") && !strings.Contains(got, "Varys") {
		t.Errorf("Expected standard log path, got %s", got)
	}
}
