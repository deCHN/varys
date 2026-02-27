package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScrape(t *testing.T) {
	// 模拟一个简单的网页服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><head><title>Test Title</title></head><body><article><h1>Main Heading</h1><p>This is the test content of the article.</p></article></body></html>`)
	}))
	defer ts.Close()

	s := NewScraper()
	art, err := s.Scrape(ts.URL)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Readability could pick different things as title depending on internal logic
	if art.Title == "" {
		t.Errorf("Expected title, got empty")
	}

	if !strings.Contains(art.Content, "test content") {
		t.Errorf("Expected content to contain 'test content', got '%s'", art.Content)
	}
}

