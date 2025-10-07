package model

import (
	"broomie/internal/scanner"
	"broomie/internal/ui"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type scanCompleteMsg struct {
	results []*scanner.ScanResult
}

type Model struct {
	Output  []string
	results []*scanner.ScanResult

	width  int
	height int
	table  ui.FileTableModel

	refresh      key.Binding
	quit         key.Binding
	quitAndPrint key.Binding
}

func InitialModel() Model {
	return Model{
		table:        ui.NewFileTableModel(),
		refresh:      key.NewBinding(key.WithKeys("R")),
		quit:         key.NewBinding(key.WithKeys("q")),
		quitAndPrint: key.NewBinding(key.WithKeys("X")),
	}
}

func (m Model) Init() tea.Cmd {
	return scan
}

func scan() tea.Msg {
	// TODO: implement loading screen
	r := scanner.ScanForJunk()
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
		m.table.SetDimensions(m.width-2, m.height-2)

	case scanCompleteMsg:
		m.results = msg.results
		m.table.SetResults(m.results)

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
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.table.View()
}
