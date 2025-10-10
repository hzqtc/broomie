package ui

import (
	"broomie/internal/scanner"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Generated with 'figlet -f epic broomie'
const logo = `
 ______   _______  _______  _______  _______ _________ _______
(  ___ \ (  ____ )(  ___  )(  ___  )(       )\__   __/(  ____ \
| (   ) )| (    )|| (   ) || (   ) || () () |   ) (   | (    \/
| (__/ / | (____)|| |   | || |   | || || || |   | |   | (__
|  __ (  |     __)| |   | || |   | || |(_)| |   | |   |  __)
| (  \ \ | (\ (   | |   | || |   | || |   | |   | |   | (
| )___) )| ) \ \__| (___) || (___) || )   ( |___) (___| (____/\
|/ \___/ |/   \__/(_______)(_______)|/     \|\_______/(_______/

`

var (
	logoStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(highlightColor)
)

type LoadingModel struct {
	isLoading bool

	progress *scanner.ScanProgress
	spinner  spinner.Model
}

func NewLoadingModel() LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return LoadingModel{
		isLoading: true,
		spinner:   s,
		progress:  &scanner.ScanProgress{},
	}
}

func (m *LoadingModel) Progress() *scanner.ScanProgress {
	return m.progress
}

func (m *LoadingModel) StartLoading() tea.Cmd {
	m.isLoading = true
	return m.spinner.Tick
}

func (m *LoadingModel) StopLoading() {
	m.isLoading = false
}

func (m LoadingModel) Update(msg tea.Msg) (LoadingModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
		}
	}
	return m, cmd
}

func (m LoadingModel) View() string {
	if !m.isLoading {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(logoStyle.Render(logo))
	sb.WriteString("\n\n")
	totalNumTasks := len(m.progress.Tasks)
	for i, t := range m.progress.Tasks {
		sb.WriteString(fmt.Sprintf("[%d/%d] %s...", i+1, totalNumTasks, t.Description))
		if t.Completed {
			sb.WriteString(logoStyle.Render(fmt.Sprintf("done in %dms", t.Duration().Milliseconds())))
		}
		sb.WriteRune('\n')
	}
	sb.WriteRune('\n')
	sb.WriteString(fmt.Sprintf("Loading...%s", spinnerStyle.Render(m.spinner.View())))
	return sb.String()
}
