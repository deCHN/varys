package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir_Safe(t *testing.T) {
	mockHome := t.TempDir()
	
	// Mock userHomeDir to return our temp dir
	oldHomeFunc := userHomeDir
	userHomeDir = func() (string, error) { return mockHome, nil }
	defer func() { userHomeDir = oldHomeFunc }()

	// 1. Test Environment Variable Priority
	t.Run("EnvVarPriority", func(t *testing.T) {
		expected := filepath.Join(mockHome, "custom_env_config")
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

	// 2. Test XDG/CLI habit (Mocked)
	t.Run("XDGHabitPriority", func(t *testing.T) {
		os.Unsetenv("VARYS_CONFIG_DIR")
		
		xdgDir := filepath.Join(mockHome, ".config", "varys")
		os.MkdirAll(xdgDir, 0755)
		configPath := filepath.Join(xdgDir, "config.json")
		os.WriteFile(configPath, []byte("{}"), 0644)

		got, err := GetConfigDir()
		if err != nil {
			t.Errorf("GetConfigDir failed: %v", err)
		}
		if got != xdgDir {
			t.Errorf("Expected %s, got %s", xdgDir, got)
		}
	})
}

func TestGetLogDir_Safe(t *testing.T) {
	mockHome := t.TempDir()
	
	// Mock userHomeDir
	oldHomeFunc := userHomeDir
	userHomeDir = func() (string, error) { return mockHome, nil }
	defer func() { userHomeDir = oldHomeFunc }()

	// 模拟 Library/Logs 存在 (macOS 场景)
	macLogPath := filepath.Join(mockHome, "Library", "Logs")
	os.MkdirAll(macLogPath, 0755)

	got, err := GetLogDir()
	if err != nil {
		t.Errorf("GetLogDir failed: %v", err)
	}

	expected := filepath.Join(mockHome, "Library", "Logs", "Varys")
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}
