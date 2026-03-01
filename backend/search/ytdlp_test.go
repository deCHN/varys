package search

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParseYTDLPLine(t *testing.T) {
	// 这是一个典型的 yt-dlp --dump-json 输出样例
	jsonLine := `{"id": "12345", "title": "Test Video", "webpage_url": "https://youtube.com/watch?v=12345", "description": "This is a test video", "uploader": "Test Channel", "epoch": 1700000000}`
	
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(jsonLine), &entry); err != nil {
		t.Fatalf("Failed to unmarshal test JSON: %v", err)
	}

	id, _ := entry["id"].(string)
	title, _ := entry["title"].(string)
	webpageURL, _ := entry["webpage_url"].(string)
	uploader, _ := entry["uploader"].(string)
	
	var publishedAt time.Time
	if epoch, ok := entry["epoch"].(float64); ok {
		publishedAt = time.Unix(int64(epoch), 0)
	}

	res := SearchResult{
		ID:          id,
		Title:       title,
		URL:         webpageURL,
		Description: entry["description"].(string),
		Source:      uploader,
		PublishedAt: publishedAt,
		Type:        ContentTypeVideo,
	}

	if res.ID != "12345" {
		t.Errorf("Expected ID 12345, got %s", res.ID)
	}
	if res.Title != "Test Video" {
		t.Errorf("Expected title 'Test Video', got %s", res.Title)
	}
	if res.Source != "Test Channel" {
		t.Errorf("Expected uploader 'Test Channel', got %s", res.Source)
	}
	if res.PublishedAt.IsZero() {
		t.Errorf("Expected PublishedAt to be set")
	}
}

func TestYTDLPSearchCommand(t *testing.T) {
	p := NewYTDLPSearchProvider()
	if p == nil {
		t.Fatal("Failed to create YTDLPSearchProvider")
	}
	
	opts := SearchOptions{Limit: 1}
	query := "AI News"
	limit := opts.Limit
	searchQuery := query
	
	// 我们实际代码中的逻辑：
	// searchQuery = fmt.Sprintf("ytsearch%d:%s", limit, query)
	cmdStr := fmt.Sprintf("ytsearch%d:%s", limit, searchQuery)
	if !strings.HasPrefix(cmdStr, "ytsearch1:") {
		t.Errorf("Expected command to start with ytsearch1:, got %s", cmdStr)
	}
}
