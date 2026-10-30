package tui

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

type (
	column struct {
		title  string
		max    int // widest the column grows from its content, 0 for unlimited
		expand int // share of the width left over after content
	}

	cell struct {
		text  string
		color color.Color // nil for the default color
	}

	// table is a scrollable list of rows with a selected row.
	table struct {
		cols      []column
		rows      [][]cell
		cursor    int
		offset    int
		vis       int   // rows shown by the last draw
		colWidths []int // column widths of the last draw
	}
)

func (t *table) set(cols []column, rows [][]cell) {
	t.cols, t.rows = cols, rows
	t.cursor = min(t.cursor, max(len(rows)-1, 0))
}

func (t *table) selected() (int, bool) {
	if t.cursor < 0 || t.cursor >= len(t.rows) {
		return 0, false
	}
	return t.cursor, true
}

func (t *table) selectRow(i int) { t.cursor = min(max(i, 0), max(len(t.rows)-1, 0)) }

func (t *table) scroll(n int) { t.selectRow(t.cursor + n) }

// key moves the selection, reporting whether the key was a movement.
func (t *table) key(key string) bool {
	page := max(t.vis-1, 1)
	switch key {
	case "up", "k":
		t.scroll(-1)
	case "down", "j":
		t.scroll(1)
	case "pgup":
		t.scroll(-page)
	case "pgdown":
		t.scroll(page)
	case "home", "g":
		t.selectRow(0)
	case "end", "G":
		t.selectRow(len(t.rows) - 1)
	default:
		return false
	}
	return true
}

// widths lays out the columns in w cells.
func (t *table) widths(w int) []int {
	const gap = 2
	widths := make([]int, len(t.cols))
	expand := 0
	total := 1 + gap*(len(t.cols)-1)
	for i, col := range t.cols {
		n := ansi.StringWidth(col.title)
		for _, row := range t.rows {
			n = max(n, ansi.StringWidth(row[i].text))
		}
		if col.max > 0 {
			n = min(n, col.max)
		}
		widths[i] = n
		expand += col.expand
		total += n
	}

	if extra := w - total; extra > 0 && expand > 0 {
		for i, col := range t.cols {
			widths[i] += extra * col.expand / expand
		}
	}
	// shrink the widest column, preferring those that expand
	for over := total - w; over > 0; over-- {
		widest := -1
		for i, col := range t.cols {
			if widths[i] <= 4 {
				continue
			}
			if widest < 0 || (col.expand > 0) != (t.cols[widest].expand > 0) && col.expand > 0 ||
				(col.expand > 0) == (t.cols[widest].expand > 0) && widths[i] > widths[widest] {
				widest = i
			}
		}
		if widest < 0 {
			break
		}
		widths[widest]--
	}
	return widths
}

// view draws the table with a header row in w by h cells, highlighting the
// selected row when focused.
func (t *table) view(w, h int, focused bool) string {
	t.vis = max(h-1, 0)
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+t.vis {
		t.offset = t.cursor - t.vis + 1
	}
	t.offset = max(t.offset, 0)

	widths := t.widths(w)
	t.colWidths = widths
	render := func(row []cell, selected bool) string {
		var line strings.Builder
		line.WriteString(" ")
		for i, c := range row {
			text := fit(c.text, widths[i])
			if !selected && c.color != nil {
				text = fg(c.color).Render(text)
			}
			line.WriteString(text)
			if i < len(row)-1 {
				line.WriteString("  ")
			}
		}
		if selected {
			return styleSelected.Render(fit(line.String(), w))
		}
		return line.String()
	}

	header := make([]cell, len(t.cols))
	for i, col := range t.cols {
		header[i] = cell{text: col.title}
	}
	lines := []string{styleHeader.Render(fit(" "+joinHeader(header, widths), w))}
	for i := t.offset; i < len(t.rows) && i < t.offset+t.vis; i++ {
		lines = append(lines, render(t.rows[i], focused && i == t.cursor))
	}
	return strings.Join(lines, "\n")
}

// columnAt returns the column under x, relative to the left edge of the drawn
// table, or -1.
func (t *table) columnAt(x int) int {
	pos := 1
	for i, w := range t.colWidths {
		if x >= pos && x < pos+w {
			return i
		}
		pos += w + 2
	}
	return -1
}

// rowAt returns the row under y, relative to the top of the drawn table
// including the header, or -1.
func (t *table) rowAt(y int) int {
	i := t.offset + y - 1
	if y < 1 || i >= len(t.rows) || i >= t.offset+t.vis {
		return -1
	}
	return i
}

func joinHeader(header []cell, widths []int) string {
	parts := make([]string, len(header))
	for i, c := range header {
		parts[i] = fit(c.text, widths[i])
	}
	return strings.Join(parts, "  ")
}
