package main

import (
	"fmt"
	"strings"

	"Varys/backend/search"

	"github.com/charmbracelet/bubbles/key"
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
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9D00FF")).
				Bold(true)
)

type item struct {
	result   search.SearchResult
	selected bool
	title    string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return fmt.Sprintf("[%s] %s", i.result.Source, i.result.URL) }
func (i item) FilterValue() string { return i.result.Title + " " + i.result.Description }

type SearchModel struct {
	list         list.Model
	results      []search.SearchResult
	selected     map[int]bool
	quitting     bool
	choice       []search.SearchResult
	query        string
	providerName string
}

func NewSearchModel(query string, providerName string, results []search.SearchResult) SearchModel {
	items := make([]list.Item, len(results))
	for i, res := range results {
		items[i] = item{result: res, selected: false, title: res.Title}
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lipgloss.Color("#9D00FF"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lipgloss.Color("#BC8F8F"))

	l := list.New(items, d, 0, 0)
	l.Title = fmt.Sprintf("Search results for: %s (via %s)", query, providerName)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	
	// 设置多选提示
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "select")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		}
	}

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
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			var finalChoices []search.SearchResult
			for idx, selected := range m.selected {
				if selected && idx < len(m.results) {
					finalChoices = append(finalChoices, m.results[idx])
				}
			}

			if len(finalChoices) == 0 {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					finalChoices = append(finalChoices, i.result)
				}
			}

			m.choice = finalChoices
			return m, tea.Quit

		case " ":
			idx := m.list.Index()
			m.selected[idx] = !m.selected[idx]
			
			items := m.list.Items()
			if idx < len(items) {
				curr := items[idx].(item)
				curr.selected = m.selected[idx]
				
				// 视觉反馈：在标题前增加 [x]
				if curr.selected {
					if !strings.HasPrefix(curr.title, "[x] ") {
						curr.title = "[x] " + curr.title
					}
				} else {
					curr.title = m.results[idx].Title
				}
				
				items[idx] = curr
				m.list.SetItems(items)
			}
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

	res, ok := finalModel.(SearchModel)
	if !ok {
		return nil, fmt.Errorf("invalid model returned")
	}
	return res.choice, nil
}
