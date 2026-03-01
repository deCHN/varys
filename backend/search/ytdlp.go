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

	// 构造 yt-dlp 搜索参数
	// ytsearchdateN:query 用于按日期排序
	searchQuery := fmt.Sprintf("ytsearch%d:%s", limit, query)
	if opts.TimeRange != "" {
		// 实际上 ytsearchdate 是针对 YouTube 的按日期排序
		searchQuery = fmt.Sprintf("ytsearchdate%d:%s", limit, query)
	}

	args := []string{
		"--dump-json",
		"--flat-playlist",
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
		if epoch, ok := entry["epoch"].(float64); ok {
			publishedAt = time.Unix(int64(epoch), 0)
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
