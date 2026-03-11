package search

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type YTDLPSearchProvider struct {
	Name string
}

func NewYTDLPSearchProvider() *YTDLPSearchProvider {
	return &YTDLPSearchProvider{
		Name: "yt-dlp",
	}
}

func (p *YTDLPSearchProvider) GetName() string {
	return p.Name
}

func (p *YTDLPSearchProvider) Search(query string, opts SearchOptions) ([]SearchResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	// Construct yt-dlp search parameters
	// ytsearchdateN:query is used for sorting by date
	searchQuery := fmt.Sprintf("ytsearch%d:%s", limit, query)
	if opts.TimeRange != "" {
		// ytsearchdate is specifically for YouTube's date-based sorting
		searchQuery = fmt.Sprintf("ytsearchdate%d:%s", limit, query)
	}

	args := []string{
		"--dump-json",
		"--cookies-from-browser", "chrome",
		"--quiet",
		searchQuery,
	}

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run yt-dlp: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	results := make([]SearchResult, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		id, _ := entry["id"].(string)
		title, _ := entry["title"].(string)
		webpageURL, _ := entry["webpage_url"].(string)
		if webpageURL == "" {
			webpageURL, _ = entry["url"].(string)
		}
		description, _ := entry["description"].(string)
		uploader, _ := entry["uploader"].(string)
		
		var publishedAt time.Time
		if dateStr, ok := entry["upload_date"].(string); ok && dateStr != "" {
			if t, err := time.Parse("20060102", dateStr); err == nil {
				publishedAt = t
			}
		}

		if publishedAt.IsZero() {
			if ts, ok := entry["timestamp"].(float64); ok {
				publishedAt = time.Unix(int64(ts), 0)
			} else if rts, ok := entry["release_timestamp"].(float64); ok {
				publishedAt = time.Unix(int64(rts), 0)
			}
		}

		results = append(results, SearchResult{
			ID:          id,
			Title:       title,
			URL:         webpageURL,
			Description: description,
			Source:      uploader,
			PublishedAt: publishedAt,
			Type:        ContentTypeVideo,
		})
	}

	return results, nil
}
