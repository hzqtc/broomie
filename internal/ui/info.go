package ui

import (
	"broomie/internal/scanner"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

type InfoPanelModel struct {
	selection []*scanner.ScanResult
}

var infoPanelStyle = lipgloss.NewStyle().
	Padding(1 /* vertical */, 2 /* horizontal */)

func NewInfoPanelModel() InfoPanelModel {
	return InfoPanelModel{}
}

func (m *InfoPanelModel) SetSelection(results []*scanner.ScanResult) {
	m.selection = results
}

func (m *InfoPanelModel) SetWidth(w int) {
	infoPanelStyle = infoPanelStyle.Width(w)
}

func (m InfoPanelModel) View() string {
	var totalSize uint64
	for _, r := range m.selection {
		totalSize += r.Size
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"%s quit | %s quit and print | %s/%s change sort column | %s change sort direction | %s toggle selection | %s select all | %s unselect all\n",
		keyStyle.Render("q"),
		keyStyle.Render("X"),
		keyStyle.Render("s"),
		keyStyle.Render("S"),
		keyStyle.Render("r"),
		keyStyle.Render("<space>"),
		keyStyle.Render("a"),
		keyStyle.Render("A"),
	))
	sb.WriteString(fmt.Sprintf(
		"Selected: %s directories freeing up %s\n",
		keyStyle.Render(fmt.Sprintf("%d", len(m.selection))),
		keyStyle.Render(humanize.Bytes(totalSize)),
	))
	return infoPanelStyle.Render(sb.String())
}
