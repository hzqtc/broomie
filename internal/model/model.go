package model

import (
	"broomie/internal/scanner"
	"broomie/internal/ui"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type scanCompleteMsg struct {
	results []*scanner.ScanResult
}

type Model struct {
	Output  []string
	results []*scanner.ScanResult

	width     int
	height    int
	table     ui.FileTableModel
	selection ui.InfoPanelModel
	loading   ui.LoadingModel

	refresh      key.Binding
	quit         key.Binding
	quitAndPrint key.Binding
}

func InitialModel() Model {
	return Model{
		table:        ui.NewFileTableModel(),
		selection:    ui.NewInfoPanelModel(),
		loading:      ui.NewLoadingModel(),
		refresh:      key.NewBinding(key.WithKeys("R")),
		quit:         key.NewBinding(key.WithKeys("q")),
		quitAndPrint: key.NewBinding(key.WithKeys("X")),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.scan, m.loading.StartLoading())
}

func (m *Model) scan() tea.Msg {
	r := scanner.ScanForJunk(m.loading.Progress())
	return scanCompleteMsg{results: r}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetDimensions(m.width-2, m.height-2-lipgloss.Height(m.selection.View()))

	case spinner.TickMsg:
		m.loading, cmd = m.loading.Update(msg)
		cmds = append(cmds, cmd)

	case scanCompleteMsg:
		m.results = msg.results
		m.table.SetResults(m.results)
		m.loading.StopLoading()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.quit):
			return m, tea.Quit
		case key.Matches(msg, m.quitAndPrint):
			for _, r := range m.table.SelectedResults() {
				m.Output = append(m.Output, r.Path)
			}
			return m, tea.Quit
		default:
			m.table, cmd = m.table.Update(msg)
			m.selection.SetSelection(m.table.SelectedResults())
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if loading := m.loading.View(); loading != "" {
		return loading
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), m.selection.View())
}
