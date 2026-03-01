package search

import "time"

// ContentType 定义搜索内容的类型
type ContentType string

const (
	ContentTypeVideo   ContentType = "video"
	ContentTypeArticle ContentType = "article"
	ContentTypeAudio   ContentType = "audio"
	ContentTypeAll     ContentType = "all"
)

// SearchOptions 提供通用的搜索参数
type SearchOptions struct {
	Limit     int
	TimeRange string // "day", "week", "month", "year"
	Type      ContentType
}

// SearchResult 标准化的搜索结果格式
type SearchResult struct {
	ID          string      `json:"id" yaml:"id"`
	Title       string      `json:"title" yaml:"title"`
	URL         string      `json:"url" yaml:"url"`
	Description string      `json:"description" yaml:"description"`
	Source      string      `json:"source" yaml:"source"`
	PublishedAt time.Time   `json:"published_at" yaml:"published_at"`
	Type        ContentType `json:"type" yaml:"type"`
}

// SearchProvider 搜索服务提供者必须实现的接口
type SearchProvider interface {
	GetName() string
	Search(query string, opts SearchOptions) ([]SearchResult, error)
}
