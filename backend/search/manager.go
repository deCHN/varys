package search

import "fmt"

type SearchManager struct {
	providers map[string]SearchProvider
}

func NewSearchManager(tavilyAPIKey string) *SearchManager {
	m := &SearchManager{
		providers: make(map[string]SearchProvider),
	}
	m.providers["yt-dlp"] = NewYTDLPSearchProvider()
	if tavilyAPIKey != "" {
		m.providers["tavily"] = NewTavilySearchProvider(tavilyAPIKey)
	}
	return m
}

func (m *SearchManager) GetProvider(name string) (SearchProvider, error) {
	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("search provider %s not found", name)
	}
	return p, nil
}

func (m *SearchManager) ListProviders() []string {
	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}
