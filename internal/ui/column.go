package ui

import (
	"broomie/internal/scanner"

	"github.com/dustin/go-humanize"
)

type column int

const (
	colSelected column = iota
	colPath
	colSize
	colDate
	colReason

	totalNumColumns
)

var colWidthMap = map[column]int{
	colSelected: 1,
	colPath:     30,
	colSize:     8,
	colDate:     15,
	colReason:   20,
}

func (c column) String() string {
	switch c {
	case colSelected:
		return " "
	case colPath:
		return "Path"
	case colSize:
		return "Size"
	case colDate:
		return "Date"
	case colReason:
		return "Reason"
	default:
		return "Unknown"
	}
}

func (c column) sortable() bool {
	return c != colSelected
}

func (c column) rightAligned() bool {
	return c == colSize
}

func (c column) width() int {
	return colWidthMap[c]
}

func (c column) nextColumn() column {
	return column((int(c) + 1) % int(totalNumColumns))
}

func (c column) prevColumn() column {
	return column((int(c) - 1 + int(totalNumColumns)) % int(totalNumColumns))
}

func (c column) getColumnData(sr *scanner.ScanResult) string {
	switch c {
	case colSelected:
		if sr.Selected {
			return ""
		} else {
			return ""
		}
	case colPath:
		return sr.Path
	case colSize:
		return humanize.Bytes(sr.Size)
	case colDate:
		return humanize.Time(sr.ModifiedDate)
	case colReason:
		return string(sr.Reason)
	default:
		return ""
	}
}
