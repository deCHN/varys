package search

import "time"

// ContentType defines the type of content being searched
type ContentType string

const (
	ContentTypeVideo   ContentType = "video"
	ContentTypeArticle ContentType = "article"
	ContentTypeAudio   ContentType = "audio"
	ContentTypeAll     ContentType = "all"
)

// SearchOptions provides common parameters for search requests
type SearchOptions struct {
	Limit     int
	TimeRange string // "day", "week", "month", "year"
	Type      ContentType
}

// SearchResult represents a standardized format for search results
type SearchResult struct {
	ID          string      `json:"id" yaml:"id"`
	Title       string      `json:"title" yaml:"title"`
	URL         string      `json:"url" yaml:"url"`
	Description string      `json:"description" yaml:"description"`
	Source      string      `json:"source" yaml:"source"`
	PublishedAt time.Time   `json:"published_at" yaml:"published_at"`
	Type        ContentType `json:"type" yaml:"type"`
}

// SearchProvider is the interface that search service providers must implement
type SearchProvider interface {
	GetName() string
	Search(query string, opts SearchOptions) ([]SearchResult, error)
}
