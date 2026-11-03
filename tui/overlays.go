package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type (
	// confirmDialog is a yes/no modal.
	confirmDialog struct {
		app  *App
		text string
		onOK func()
		ok   bool // OK button selected, Cancel is the default

		// hit data of the last draw, relative to the dialog
		width, height int
		buttonY       int
		cancelX, okX  [2]int
	}

	// picker is a searchable single choice list.
	picker struct {
		app      *App
		title    string
		items    []string
		onSelect func(string)
		filter   textinput.Model
		matches  []string
		width    int
		height   int
		table    table
	}

	// filterDialog is a checkbox list of log levels to show, applied live.
	filterDialog struct {
		app      *App
		filters  []logFilter
		onChange func()
		cursor   int
		width    int
	}

	// textDialog is a scrollable, dismissible text page.
	textDialog struct {
		app   *App
		title string
		vp    viewport.Model
	}
)

// confirm shows a yes/no modal, running onOK when confirmed.
func (a *App) confirm(text string, onOK func()) {
	a.openOverlay(&confirmDialog{app: a, text: text, onOK: onOK})
}

func (c *confirmDialog) floating() bool { return true }

func (c *confirmDialog) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		c.click(click)
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}
	switch key.String() {
	case "left", "right", "tab", "shift+tab", "h", "l":
		c.ok = !c.ok
	case "y":
		c.choose(true)
	case "n", "esc":
		c.choose(false)
	case "enter":
		c.choose(c.ok)
	}
}

func (c *confirmDialog) click(click tea.MouseClickMsg) {
	x, y := click.X-c.app.overlayX, click.Y-c.app.overlayY
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		c.choose(false)
		return
	}
	if click.Button != tea.MouseLeft || y != c.buttonY {
		return
	}
	switch {
	case x >= c.cancelX[0] && x < c.cancelX[1]:
		c.choose(false)
	case x >= c.okX[0] && x < c.okX[1]:
		c.choose(true)
	}
}

func (c *confirmDialog) choose(ok bool) {
	c.app.closeOverlay(c)
	if ok {
		c.onOK()
	}
}

func (c *confirmDialog) view(w, _ int) string {
	width := min(max(w/3, 40), w-4)
	text := lipgloss.NewStyle().Width(width - 4).Align(lipgloss.Center).Render(c.text)

	button := func(label string, selected bool) string {
		style := lipgloss.NewStyle().Padding(0, 2).Background(colorDisabled).Foreground(colorValue)
		if selected {
			style = styleSelected.Padding(0, 2)
		}
		return style.Render(label)
	}
	cancel, ok := button("Cancel", !c.ok), button("OK", c.ok)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, cancel, "  ", ok)
	content := text + "\n\n" + lipgloss.PlaceHorizontal(width-2, lipgloss.Center, buttons)

	c.width, c.height = width, lipgloss.Height(content)+4
	c.buttonY = lipgloss.Height(content) + 1
	x := 1 + (width-2-lipgloss.Width(buttons))/2
	c.cancelX = [2]int{x, x + lipgloss.Width(cancel)}
	c.okX = [2]int{c.cancelX[1] + 2, x + lipgloss.Width(buttons)}
	return box("", "\n"+content, width, c.height, colorValue)
}

// pick shows a searchable single choice list.
func (a *App) pick(title string, items []string, onSelect func(string)) {
	p := &picker{app: a, title: title, items: items, onSelect: onSelect, filter: newInput("")}
	p.refilter()
	a.openOverlay(p)
}

func (p *picker) floating() bool { return true }

func (p *picker) refilter() {
	needle := strings.ToLower(p.filter.Value())
	p.matches = p.matches[:0]
	rows := [][]cell{}
	for _, item := range p.items {
		if needle == "" || strings.Contains(strings.ToLower(item), needle) {
			p.matches = append(p.matches, item)
			rows = append(rows, []cell{{text: item}})
		}
	}
	p.table.set([]column{{title: "", expand: 1}}, rows)
	p.table.selectRow(0)
}

func (p *picker) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		// the list starts below the border and the filter row
		row := p.table.offset + click.Y - p.app.overlayY - 2
		x, y := click.X-p.app.overlayX, click.Y-p.app.overlayY
		if x < 0 || x >= p.width || y < 0 || y >= p.height {
			p.app.closeOverlay(p)
			return
		}
		if click.Button == tea.MouseLeft && x > 0 && x < p.width-1 && y >= 2 && row < len(p.matches) && row < p.table.offset+p.table.vis-1 {
			choice := p.matches[row]
			p.app.closeOverlay(p)
			p.onSelect(choice)
		}
		return
	}
	if delta := wheelDelta(msg); delta != 0 {
		p.table.scroll(delta)
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		p.filter, _ = p.filter.Update(msg)
		p.refilter()
		return
	}
	switch key.String() {
	case "esc":
		p.app.closeOverlay(p)
	case "enter":
		if i, ok := p.table.selected(); ok {
			choice := p.matches[i]
			p.app.closeOverlay(p)
			p.onSelect(choice)
		}
	case "up", "down", "pgup", "pgdown", "home", "end":
		p.table.key(key.String())
	default:
		before := p.filter.Value()
		p.filter, _ = p.filter.Update(msg)
		if p.filter.Value() != before {
			p.refilter()
		}
	}
}

func (p *picker) view(w, h int) string {
	width, height := min(60, w-4), min(20, h)
	p.width, p.height = width, height
	p.filter.SetWidth(width - 6)
	content := " " + styleTitle.Render("/") + " " + p.filter.View() + "\n" + p.listView(width-2, height-4) +
		"\n " + hint("enter", "select") + "  " + hint("esc", "close")
	return box(" "+p.title+" ", content, width, height, colorValue)
}

// listView draws the matches without a header row.
func (p *picker) listView(w, h int) string {
	vis := max(h, 0)
	p.table.vis = vis + 1
	if p.table.cursor < p.table.offset {
		p.table.offset = p.table.cursor
	}
	if p.table.cursor >= p.table.offset+vis {
		p.table.offset = p.table.cursor - vis + 1
	}
	p.table.offset = max(p.table.offset, 0)

	var lines []string
	for i := p.table.offset; i < len(p.matches) && i < p.table.offset+vis; i++ {
		line := fit(" "+p.matches[i], w)
		if i == p.table.cursor {
			line = styleSelected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// showText shows a scrollable, dismissible text page.
func (a *App) showText(title, text string) { a.showStyledText(title, text) }

func (a *App) showStyledText(title, text string) {
	vp := viewport.New()
	vp.SetContent(text)
	a.openOverlay(&textDialog{app: a, title: title, vp: vp})
}

func (t *textDialog) floating() bool { return false }

func (t *textDialog) update(msg tea.Msg) {
	if delta := wheelDelta(msg); delta != 0 {
		if delta < 0 {
			t.vp.ScrollUp(-delta)
		} else {
			t.vp.ScrollDown(delta)
		}
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}
	switch key.String() {
	case "esc", "q", "?":
		t.app.closeOverlay(t)
	default:
		scrollKey(&t.vp, key.String())
	}
}

func (t *textDialog) view(w, h int) string {
	t.vp.SetWidth(w - 2)
	t.vp.SetHeight(h - 2)
	return box(" "+t.title+" (esc to close) ", t.vp.View(), w, h, colorValue)
}

// scrollKey applies pager keys to a viewport, reporting whether it was one.
func scrollKey(vp *viewport.Model, key string) bool {
	switch key {
	case "up", "k":
		vp.ScrollUp(1)
	case "down", "j":
		vp.ScrollDown(1)
	case "pgup":
		vp.PageUp()
	case "pgdown", "space":
		vp.PageDown()
	case "home", "g":
		vp.GotoTop()
	case "end", "G":
		vp.GotoBottom()
	case "left", "h":
		vp.ScrollLeft(4)
	case "right", "l":
		vp.ScrollRight(4)
	default:
		return false
	}
	return true
}

func (d *filterDialog) floating() bool { return true }

func (d *filterDialog) toggle(i int) {
	d.cursor = i
	d.filters[i].hidden = !d.filters[i].hidden
	d.onChange()
}

func (d *filterDialog) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		x, y := click.X-d.app.overlayX, click.Y-d.app.overlayY
		switch {
		case x < 0 || x >= d.width || y < 0 || y >= len(d.filters)+4:
			d.app.closeOverlay(d)
		case click.Button == tea.MouseLeft && y >= 1 && y <= len(d.filters):
			d.toggle(y - 1)
		}
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}
	switch key.String() {
	case "esc", "q", "f", "enter":
		d.app.closeOverlay(d)
	case "up", "k":
		d.cursor = max(d.cursor-1, 0)
	case "down", "j":
		d.cursor = min(d.cursor+1, len(d.filters)-1)
	case "space", "x":
		d.toggle(d.cursor)
	}
}

func (d *filterDialog) view(w, _ int) string {
	d.width = min(34, w-4)
	lines := make([]string, len(d.filters))
	for i, f := range d.filters {
		mark := "[x]"
		if f.hidden {
			mark = "[ ]"
		}
		line := fit(" "+mark+" "+f.label, d.width-2)
		if i == d.cursor {
			line = styleSelected.Render(line)
		}
		lines[i] = line
	}
	content := strings.Join(lines, "\n") + "\n\n " + hint("space", "toggle") + "  " + hint("esc", "close")
	return box(" Filter log ", content, d.width, len(d.filters)+4, colorValue)
}
