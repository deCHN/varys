package main

import (
	"fmt"

	"Varys/backend/search"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#290137")).
			Padding(0, 1)
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

type item struct {
	result search.SearchResult
}

func (i item) Title() string       { return i.result.Title }
func (i item) Description() string { return fmt.Sprintf("[%s] %s", i.result.Source, i.result.URL) }
func (i item) FilterValue() string { return i.result.Title + " " + i.result.Description }

type SearchModel struct {
	list         list.Model
	results      []search.SearchResult
	selected     map[int]bool
	quitting     bool
	choice       []search.SearchResult
	err          error
	loading      bool
	query        string
	providerName string
}

func NewSearchModel(query string, providerName string, results []search.SearchResult) SearchModel {
	items := make([]list.Item, len(results))
	for i, res := range results {
		items[i] = item{result: res}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("Search results for: %s (via %s)", query, providerName)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return SearchModel{
		list:         l,
		results:      results,
		selected:     make(map[int]bool),
		query:        query,
		providerName: providerName,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return nil
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = append(m.choice, i.result)
				return m, tea.Quit
			}

		case " ":
			idx := m.list.Index()
			m.selected[idx] = !m.selected[idx]
			return m, nil
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SearchModel) View() string {
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View())
}

// RunSearchTUI 启动搜索界面的入口函数
func RunSearchTUI(query string, provider search.SearchProvider, opts search.SearchOptions) ([]search.SearchResult, error) {
	fmt.Printf("Searching for '%s' using %s...\n", query, provider.GetName())
	
	results, err := provider.Search(query, opts)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	m := NewSearchModel(query, provider.GetName(), results)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	res := finalModel.(SearchModel)
	return res.choice, nil
}
