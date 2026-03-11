package search

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
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

	ytPath := "yt-dlp"
	cmd := exec.Command(ytPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	results := make([]SearchResult, 0, limit)
	scanner := bufio.NewScanner(stdout)
	// Use 1MB buffer to handle large JSON records in detailed mode (lots of formats/subtitles)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	
	for scanner.Scan() {
		line := scanner.Text()
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

		if opts.OnProgress != nil {
			opts.OnProgress(len(results), limit)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error during search: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("yt-dlp search command failed: %w", err)
	}

	return results, nil
}
