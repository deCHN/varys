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

	// 2. Verify key phrases from our recent update exist
	expectedPhrase := "First, summarize the full text clearly in Simplified Chinese"
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
