package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	mgr := NewManager("/tmp")

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello_World"},
		{"Test: Name", "Test_Name"},
		{"Invalid/Char?", "InvalidChar"},
		{"[Obsidian] Link", "Obsidian_Link"},
		{"   Trim Me   ", "Trim_Me"},
		{"Repeated__Underscores", "Repeated_Underscores"},
		{strings.Repeat("a", 100), strings.Repeat("a", 80)}, // Truncate
	}

	for _, tt := range tests {
		result := mgr.SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSaveNote(t *testing.T) {
	// Setup temp vault
	vaultDir, err := os.MkdirTemp("", "vault_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(vaultDir)

	mgr := NewManager(vaultDir)

	data := NoteData{
		Title:        "Test Note",
		URL:          "http://example.com",
		Language:     "en",
		Summary:      "This is a summary.",
		KeyPoints:    []string{"Point 1", "Point 2"},
		Tags:         []string{"AI", "Test"},
		Assessment:   map[string]string{"authenticity": "High"},
		OriginalText: "Original content",
		AssetsFolder: "assets",
		AudioFile:    "Test_Note.m4a",
		CreatedTime:  "2023-01-01 12:00",
	}

	// Test SaveNote
	path, err := mgr.SaveNote(data)
	if err != nil {
		t.Fatalf("SaveNote failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Note file not found at %s", path)
	}

	// Read content
	contentBytes, _ := os.ReadFile(path)
	content := string(contentBytes)

	// Verify content parts
	if !strings.Contains(content, "# Test Note") {
		t.Error("Title not found in markdown")
	}
	if !strings.Contains(content, "This is a summary.") {
		t.Error("Summary not found")
	}
	if !strings.Contains(content, "![[assets/Test_Note.m4a]]") {
		t.Error("Audio link not found")
	}
}

func TestMoveMedia(t *testing.T) {
	// Setup temp vault and source dir
	tempDir, _ := os.MkdirTemp("", "source")
	defer os.RemoveAll(tempDir)
	vaultDir, _ := os.MkdirTemp("", "vault")
	defer os.RemoveAll(vaultDir)

	// Create dummy audio
	sourceAudio := filepath.Join(tempDir, "temp.m4a")
	os.WriteFile(sourceAudio, []byte("fake audio"), 0644)

	mgr := NewManager(vaultDir)
	
	finalName, err := mgr.MoveMedia(sourceAudio, "Final_Name")
	if err != nil {
		t.Fatalf("MoveMedia failed: %v", err)
	}

	if finalName != "Final_Name.m4a" {
		t.Errorf("Expected filename Final_Name.m4a, got %s", finalName)
	}

	// Check if file exists in vault/assets
	expectedPath := filepath.Join(vaultDir, "assets", "Final_Name.m4a")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("Audio file not moved to assets folder")
	}

	// Check if source is gone
	if _, err := os.Stat(sourceAudio); err == nil {
		t.Error("Source file still exists")
	}
}
