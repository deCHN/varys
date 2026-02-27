package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type TavilySearchProvider struct {
	APIKey string
}

func NewTavilySearchProvider(apiKey string) *TavilySearchProvider {
	return &TavilySearchProvider{
		APIKey: apiKey,
	}
}

func (p *TavilySearchProvider) GetName() string {
	return "tavily"
}

type TavilyRequest struct {
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth"` // "basic" or "advanced"
	IncludeImages     bool     `json:"include_images"`
	IncludeAnswer     bool     `json:"include_answer"`
	IncludeRawContent bool     `json:"include_raw_content"`
	MaxResults        int      `json:"max_results"`
	TimeRange         string   `json:"time_range,omitempty"` // "day", "week", "month", "year"
}

type TavilyResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	Published  string  `json:"published_date,omitempty"`
}

type TavilyResponse struct {
	Results []TavilyResult `json:"results"`
}

func (p *TavilySearchProvider) Search(query string, opts SearchOptions) ([]SearchResult, error) {
	if p.APIKey == "" {
		p.APIKey = os.Getenv("TAVILY_API_KEY")
	}
	if p.APIKey == "" {
		return nil, fmt.Errorf("TAVILY_API_KEY is not set")
	}

	reqBody := TavilyRequest{
		Query:             query,
		SearchDepth:       "advanced",
		MaxResults:        opts.Limit,
		TimeRange:         opts.TimeRange,
		IncludeRawContent: false,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily api error: %s", resp.Status)
	}

	var tavilyResp TavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(tavilyResp.Results))
	for _, tr := range tavilyResp.Results {
		var publishedAt time.Time
		if tr.Published != "" {
			publishedAt, _ = time.Parse("2006-01-02", tr.Published)
		}

		results = append(results, SearchResult{
			Title:       tr.Title,
			URL:         tr.URL,
			Description: tr.Content,
			PublishedAt: publishedAt,
			Source:      "web",
			Type:        ContentTypeArticle,
		})
	}

	return results, nil
}
