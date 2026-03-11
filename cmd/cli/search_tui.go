package main

import (
	"fmt"
	"strings"

	"Varys/backend/search"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#290137")).
			Padding(0, 1)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1)
)

type item struct {
	result   search.SearchResult
	selected bool
	title    string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string {
	dateStr := "N/A"
	if !i.result.PublishedAt.IsZero() {
		dateStr = i.result.PublishedAt.Format("2006-01-02")
	}
	return fmt.Sprintf("[%s] %s | %s", i.result.Source, dateStr, i.result.URL)
}
func (i item) FilterValue() string { return i.result.Title + " " + i.result.Description }

type SearchProgressMsg struct {
	Current int
	Total   int
}

type SearchDoneMsg struct {
	Results []search.SearchResult
	Err     error
}

type SearchModel struct {
	list       list.Model
	progress   progress.Model
	spinner    spinner.Model
	results    []search.SearchResult
	selected   map[int]bool
	quitting   bool
	loading    bool
	choice     []search.SearchResult
	query      string
	provider   search.SearchProvider
	opts       search.SearchOptions
	err        error
	current    int
	total      int
	lastWidth  int
	lastHeight int
}

func NewSearchModel(query string, provider search.SearchProvider, opts search.SearchOptions) SearchModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D00FF"))

	p := progress.New(progress.WithDefaultGradient())

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("Search results for: %s", query)
	l.SetShowStatusBar(true)
	l.Styles.Title = titleStyle

	return SearchModel{
		list:     l,
		progress: p,
		spinner:  s,
		selected: make(map[int]bool),
		query:    query,
		provider: provider,
		opts:     opts,
		loading:  true,
		total:    opts.Limit,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			res, err := m.provider.Search(m.query, m.opts)
			return SearchDoneMsg{Results: res, Err: err}
		},
	)
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

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

	case SearchProgressMsg:
		m.current = msg.Current
		m.total = msg.Total
		pct := float64(m.current) / float64(m.total)
		cmds = append(cmds, m.progress.SetPercent(pct))

	case SearchDoneMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, tea.Quit
		}
		m.results = msg.Results
		items := make([]list.Item, len(m.results))
		for i, res := range m.results {
			items[i] = item{result: res, selected: false, title: res.Title}
		}
		
		d := list.NewDefaultDelegate()
		d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lipgloss.Color("#9D00FF"))
		d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lipgloss.Color("#BC8F8F"))
		m.list.SetDelegate(d)
		m.list.SetItems(items)
		m.list.Title = fmt.Sprintf("Search results for: %s (via %s)", m.query, m.provider.GetName())
		
		// Ensure list has the correct size when results arrive
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(m.lastWidth-h, m.lastHeight-v)
		
		m.list.AdditionalFullHelpKeys = func() []key.Binding {
			return []key.Binding{
				key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "select")),
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newProgressModel, ok := newModel.(progress.Model); ok {
			m.progress = newProgressModel
		}
		return m, cmd

	case tea.WindowSizeMsg:
		m.lastWidth = msg.Width
		m.lastHeight = msg.Height
		h, v := docStyle.GetFrameSize()
		m.progress.Width = msg.Width - h - 10
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	if !m.loading {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, tea.Batch(cmds...)
}

func (m SearchModel) View() string {
	if m.quitting {
		return ""
	}

	if m.loading {
		status := fmt.Sprintf("%s Fetching results for '%s'...", m.spinner.View(), m.query)
		return docStyle.Render(
			fmt.Sprintf(
				"\n%s\n\n%s %d/%d\n\n%s",
				status,
				m.progress.View(),
				m.current,
				m.total,
				helpStyle.Render("Press q to cancel"),
			),
		)
	}

	return docStyle.Render(m.list.View())
}

func RunSearchTUI(query string, provider search.SearchProvider, opts search.SearchOptions) ([]search.SearchResult, error) {
	// Create a placeholder for the program so the callback can reference it
	var p *tea.Program

	originalOnProgress := opts.OnProgress
	opts.OnProgress = func(current, total int) {
		if originalOnProgress != nil {
			originalOnProgress(current, total)
		}
		if p != nil {
			p.Send(SearchProgressMsg{Current: current, Total: total})
		}
	}

	m := NewSearchModel(query, provider, opts)
	p = tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	res, ok := finalModel.(SearchModel)
	if !ok {
		return nil, fmt.Errorf("invalid model returned")
	}

	if res.err != nil {
		return nil, res.err
	}

	return res.choice, nil
}
