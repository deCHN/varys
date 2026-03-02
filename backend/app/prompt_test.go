package app

import (
	"strings"
	"testing"

	"Varys/backend/analyzer"
)

func TestGetDefaultPrompt_Integration(t *testing.T) {
	// 1. Verify Analyzer can read the embedded file
	rawPrompt := analyzer.GetDefaultPrompt()
	if rawPrompt == "" {
		t.Fatal("analyzer.GetDefaultPrompt() returned empty string")
	}

	// 2. Verify common identity phrase exists (works for both old and new versions)
	expectedPhrase := "You are an expert content analyst"
	if !strings.Contains(rawPrompt, expectedPhrase) {
		t.Errorf("Default prompt does not contain expected phrase: %s", expectedPhrase)
	}

	// 3. Verify App bridge works
	app := NewApp()
	// Mock config manager if necessary, but GetDefaultPrompt shouldn't depend on it
	
	appPrompt := app.GetDefaultPrompt()
	if appPrompt != rawPrompt {
		t.Errorf("App.GetDefaultPrompt() mismatch. Expected length %d, got %d", len(rawPrompt), len(appPrompt))
	}
}
