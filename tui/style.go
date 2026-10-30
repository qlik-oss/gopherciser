package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// theme colors
var (
	colorTitle    = lipgloss.Color("#00ffff")
	colorKey      = lipgloss.Color("#1e90ff")
	colorValue    = lipgloss.Color("#ffffff")
	colorHeader   = lipgloss.Color("#778899")
	colorSelected = lipgloss.Color("#00ffff")
	colorError    = lipgloss.Color("#ff4500")
	colorWarning  = lipgloss.Color("#ffff00")
	colorOK       = lipgloss.Color("#32cd32")
	colorDisabled = lipgloss.Color("#808080")
	colorBlack    = lipgloss.Color("#000000")
)

var (
	styleTitle    = lipgloss.NewStyle().Foreground(colorTitle)
	styleTitleB   = styleTitle.Bold(true)
	styleKey      = lipgloss.NewStyle().Foreground(colorKey)
	styleKeyB     = styleKey.Bold(true)
	styleDisabled = lipgloss.NewStyle().Foreground(colorDisabled)
	styleSelected = lipgloss.NewStyle().Foreground(colorBlack).Background(colorSelected)
	styleHeader   = lipgloss.NewStyle().Foreground(colorHeader).Bold(true)
	styleReverse  = lipgloss.NewStyle().Reverse(true)
)

func fg(c color.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

// hint renders a key hint as in "<ctrl-s> save".
func hint(key, desc string) string { return styleKeyB.Render("<"+key+">") + " " + desc }

// fit truncates or pads a line to exactly w cells.
func fit(line string, w int) string {
	if w < 1 {
		return ""
	}
	if ansi.StringWidth(line) > w {
		line = ansi.Truncate(line, w, "…")
	}
	if pad := w - ansi.StringWidth(line); pad > 0 {
		line += strings.Repeat(" ", pad)
	}
	return line
}

// fitBlock clips or pads text to exactly w by h cells.
func fitBlock(text string, w, h int) string {
	if h < 1 {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	for i, line := range lines {
		lines[i] = fit(line, w)
	}
	return strings.Join(lines, "\n")
}

// box draws content in a single line border with the title in the top edge.
func box(title, content string, w, h int, border color.Color) string {
	iw, ih := w-2, h-2
	if iw < 1 || ih < 0 {
		return ""
	}
	edge := fg(border)
	if title != "" {
		title = ansi.Truncate(title, iw, "…")
	}
	top := edge.Render("┌") + styleTitle.Render(title) + edge.Render(strings.Repeat("─", iw-ansi.StringWidth(title))+"┐")

	lines := []string{top}
	side := edge.Render("│")
	for _, line := range strings.Split(fitBlock(content, iw, ih), "\n") {
		if ih > 0 {
			lines = append(lines, side+line+side)
		}
	}
	lines = append(lines, edge.Render("└"+strings.Repeat("─", iw)+"┘"))
	return strings.Join(lines, "\n")
}

// center places a block in the middle of a w by h area.
func center(block string, w, h int) string {
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, block)
}

// truncate shortens text to at most max runes.
func truncate(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max-1]) + "…"
}
