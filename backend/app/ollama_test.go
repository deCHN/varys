//go:build integration

package app

import (
	"Varys/backend/config"
	"Varys/backend/dependency"
	"context"
	"os/exec"
	"testing"
)

func TestApp_OllamaServiceToggle(t *testing.T) {
	if _, err := exec.LookPath("ollama"); err != nil {
		t.Skip("ollama binary not found in PATH, skipping integration test")
	}

	// Setup
	app := buildTestAppWithConfig(t, config.Config{
		AIProvider: "ollama",
	})
	
	// Inject a real dep manager
	dm, err := dependency.NewManager()
	if err != nil {
		t.Fatalf("failed to init dep manager: %v", err)
	}
	app.depManager = dm
	app.ctx = context.Background()

	// 1. Check if we can stop it (if it's running)
	_, err = app.StopOllamaService()
	if err != nil {
		t.Logf("StopOllamaService (initial) returned: %v", err)
	}

	// 2. Start it
	t.Log("Starting Ollama...")
	msg, err := app.StartOllamaService()
	if err != nil {
		t.Fatalf("Failed to start Ollama: %v", err)
	}
	t.Logf("Start msg: %s", msg)

	// 3. Verify it's running (via diagnostics)
	diag := app.GetStartupDiagnostics()
	ollamaItem := findDiagnosticItem(t, diag, "ollama")
	if ollamaItem.Status != "ok" {
		t.Errorf("Expected ollama status 'ok' after start, got '%s'", ollamaItem.Status)
	}

	// 4. Stop it
	t.Log("Stopping Ollama...")
	msg, err = app.StopOllamaService()
	if err != nil {
		t.Fatalf("Failed to stop Ollama: %v", err)
	}
	t.Logf("Stop msg: %s", msg)

	// 5. Verify it's stopped
	diag = app.GetStartupDiagnostics()
	ollamaItem = findDiagnosticItem(t, diag, "ollama")
	if ollamaItem.Status == "ok" {
		t.Error("Expected ollama status not to be 'ok' after stop")
	}
}
