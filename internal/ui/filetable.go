package ui

import (
	"fmt"
	"sort"

	"broomie/internal/scanner"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	tableStyle = baseStyle.BorderForeground(focusedBorderColor)
)

const (
	colPadding = 2
)

type sortDir int

const (
	sortAsc sortDir = iota
	sortDesc
)

var columns = func() []column {
	cols := []column{}
	for i := range totalNumColumns {
		cols = append(cols, column(i))
	}
	return cols
}()

var minTableWidth = func() int {
	minWidth := 0
	for _, c := range columns {
		minWidth += c.width() + colPadding
	}
	return minWidth
}()

type FileTableModel struct {
	// Data
	results []*scanner.ScanResult

	// UI component
	table table.Model

	// State
	sortColumn column
	sortDir

	// Key bindings
	sortNext       key.Binding
	sortPrev       key.Binding
	toggleSortDir  key.Binding
	selectAll      key.Binding
	unselectAll    key.Binding
	toggleSelected key.Binding
}

func NewFileTableModel() FileTableModel {
	m := FileTableModel{
		table: table.New(
			table.WithFocused(true),
			table.WithStyles(getTableStyles()),
		),
		sortColumn:     colReason,
		sortDir:        sortAsc,
		sortNext:       key.NewBinding(key.WithKeys("s")),
		sortPrev:       key.NewBinding(key.WithKeys("S")),
		toggleSortDir:  key.NewBinding(key.WithKeys("r")),
		selectAll:      key.NewBinding(key.WithKeys("a")),
		unselectAll:    key.NewBinding(key.WithKeys("A")),
		toggleSelected: key.NewBinding(key.WithKeys(" ")),
	}
	m.updateColumns()
	return m
}

func getTableStyles() table.Styles {
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		Foreground(highlightColor).
		BorderStyle(roundedBorder).
		BorderForeground(borderColor).
		BorderBottom(true).
		Bold(true)
	tableStyles.Selected = tableStyles.Selected.
		Foreground(highlightForegroudColor).
		Background(highlightColor).
		Bold(true)
	return tableStyles
}

func (m FileTableModel) Update(msg tea.Msg) (FileTableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.sortNext):
			m.sortNextColumn()
		case key.Matches(msg, m.sortPrev):
			m.sortPrevColumn()
		case key.Matches(msg, m.toggleSortDir):
			if m.sortDir == sortAsc {
				m.sortDir = sortDesc
			} else {
				m.sortDir = sortAsc
			}
			// Needs to update column because sorting indicator changed
			m.updateColumns()
			m.sortRows()
		case key.Matches(msg, m.selectAll):
			for _, s := range m.results {
				s.Selected = true
			}
			m.UpdateRows()
		case key.Matches(msg, m.unselectAll):
			for _, s := range m.results {
				s.Selected = false
			}
			m.UpdateRows()
		case key.Matches(msg, m.toggleSelected):
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.results) {
				r := m.results[cursor]
				r.Selected = !r.Selected
				m.UpdateRows()
			}
		default:
			m.table, _ = m.table.Update(msg)
		}
	}
	return m, nil
}

func (m FileTableModel) View() string {
	return tableStyle.Render(m.table.View())
}

func (m *FileTableModel) SetDimensions(width, height int) {
	m.table.SetWidth(width)
	m.table.SetHeight(height)
	m.updateColumns()
	m.UpdateRows()
}

func (m *FileTableModel) SetResults(results []*scanner.ScanResult) {
	m.results = results
	m.sortRows()
}

func (m *FileTableModel) SelectedResults() []*scanner.ScanResult {
	selected := []*scanner.ScanResult{}
	for _, r := range m.results {
		if r.Selected {
			selected = append(selected, r)
		}
	}
	return selected
}

func (m *FileTableModel) sortNextColumn() {
	newCol := m.sortColumn.nextColumn()
	for !newCol.sortable() {
		newCol = newCol.nextColumn()
	}
	m.sortColumn = newCol
	// Needs to update column because sorting indicator changed
	m.updateColumns()
	m.sortRows()
}

func (m *FileTableModel) sortPrevColumn() {
	newCol := m.sortColumn.prevColumn()
	for !newCol.sortable() {
		newCol = newCol.prevColumn()
	}
	m.sortColumn = newCol
	// Needs to update column because sorting indicator changed
	m.updateColumns()
	m.sortRows()
}

func (m *FileTableModel) sortRows() {
	switch m.sortColumn {
	case colSelected:
		sort.Slice(m.results, func(i, j int) bool {
			if m.sortDir == sortAsc {
				return m.results[i].Selected && !m.results[j].Selected // Selected first
			} else {
				return !m.results[i].Selected && m.results[j].Selected // Unselected first
			}
		})
	case colPath:
		sort.Slice(m.results, func(i, j int) bool {
			if m.sortDir == sortAsc {
				return m.results[i].Path < m.results[j].Path
			} else {
				return m.results[i].Path > m.results[j].Path
			}
		})
	case colDate:
		sort.Slice(m.results, func(i, j int) bool {
			if m.sortDir == sortAsc {
				return m.results[i].ModifiedDate.Before(m.results[j].ModifiedDate)
			} else {
				return m.results[i].ModifiedDate.After(m.results[j].ModifiedDate)
			}
		})
	case colSize:
		sort.Slice(m.results, func(i, j int) bool {
			if m.sortDir == sortAsc {
				return m.results[i].Size < m.results[j].Size
			} else {
				return m.results[i].Size > m.results[j].Size
			}
		})
	case colReason:
		sort.Slice(m.results, func(i, j int) bool {
			if m.sortDir == sortAsc {
				return string(m.results[i].Reason) < string(m.results[j].Reason)
			} else {
				return string(m.results[i].Reason) > string(m.results[j].Reason)
			}
		})
	}
	m.UpdateRows()
}

func (m *FileTableModel) updateColumns() {
	additionalWidth := m.table.Width() - minTableWidth
	tableCols := []table.Column{}
	for _, col := range columns {
		colTitle := col.String()
		colWidth := col.width()
		// Add sort indicator
		if col == m.sortColumn {
			if m.sortDir == sortAsc {
				colTitle = fmt.Sprintf("↑ %s", colTitle)
			} else {
				colTitle = fmt.Sprintf("↓ %s", colTitle)
			}
		}
		// Right align columns
		if col.rightAligned() {
			colTitle = fmt.Sprintf("%*s", colWidth, colTitle)
		}
		// If desc column is visible, it takes all remaining width
		if col == colPath {
			colWidth += additionalWidth
		}
		tableCols = append(tableCols, table.Column{Title: colTitle, Width: colWidth})
	}
	m.table.SetColumns(tableCols)
}

func (m *FileTableModel) UpdateRows() {
	rows := make([]table.Row, len(m.results))
	for i, r := range m.results {
		rowData := []string{}
		for _, col := range columns {
			colData := col.getColumnData(r)
			if col.rightAligned() {
				colData = fmt.Sprintf("%*s", col.width(), colData)
			}
			rowData = append(rowData, colData)
		}
		rows[i] = table.Row(rowData)
	}
	m.table.SetRows(rows)

	// Reset cursor if it's out of bounds
	if m.table.Cursor() >= len(rows) {
		m.table.SetCursor(0)
	}
}
