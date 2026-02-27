package search

import (
	"testing"
)

func TestSearchManager(t *testing.T) {
	sm := NewSearchManager("") // Empty key means tavily is disabled
	
	providers := sm.ListProviders()
	foundYTDLP := false
	for _, p := range providers {
		if p == "yt-dlp" {
			foundYTDLP = true
		}
	}
	
	if !foundYTDLP {
		t.Errorf("Expected yt-dlp provider to be registered")
	}

	p, err := sm.GetProvider("yt-dlp")
	if err != nil {
		t.Fatalf("Failed to get yt-dlp provider: %v", err)
	}

	if p.GetName() != "yt-dlp" {
		t.Errorf("Expected provider name 'yt-dlp', got '%s'", p.GetName())
	}
}
