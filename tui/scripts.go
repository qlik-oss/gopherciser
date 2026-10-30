package tui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// scriptsView lists the scenario scripts in the working directory.
type scriptsView struct {
	app    *App
	table  table
	filter textinput.Model
	// filtering is set while the filter line is shown and has focus
	filtering bool
	filterOn  bool

	scripts  []scriptInfo
	filtered []scriptInfo

	sortCol  int
	sortDesc bool
}

// sortKeys select the sorted column, the first letter of the column title.
var sortKeys = []string{"n", "a", "s", "m"}

func newScriptsView(app *App) *scriptsView {
	return &scriptsView{app: app, filter: newInput("")}
}

func (v *scriptsView) activate() {
	v.app.setHints([][2]string{
		{"enter", "open scenario"},
		{"x", "execute script"},
		{"e", "edit raw script"},
		{"n/a/s/m", "sort by column"},
		{"r", "reload list"},
		{"d", "browse folders"},
		{"/", "filter"},
		{":", "command"},
		{"?", "help"},
		{"ctrl-c", "quit"},
	})
	v.reload()
}

func (v *scriptsView) capturing() bool { return v.filtering }

func (v *scriptsView) reload() {
	scripts, err := listScripts(v.app.dir)
	if err != nil {
		v.app.flashErr(err)
		return
	}
	v.scripts = scripts
	v.render()
}

func (v *scriptsView) render() {
	needle := strings.ToLower(v.filter.Value())
	v.filtered = nil
	for _, s := range v.scripts {
		if needle == "" || strings.Contains(strings.ToLower(s.name), needle) {
			v.filtered = append(v.filtered, s)
		}
	}

	slices.SortStableFunc(v.filtered, func(a, b scriptInfo) int {
		var c int
		switch v.sortCol {
		case 1:
			c = cmp.Compare(a.actions, b.actions)
		case 2:
			c = cmp.Compare(a.size, b.size)
		case 3:
			c = a.modified.Compare(b.modified)
		default:
			c = cmp.Compare(a.name, b.name)
		}
		if v.sortDesc {
			return -c
		}
		return c
	})

	rows := make([][]cell, len(v.filtered))
	for i, s := range v.filtered {
		status := cell{text: "ok", color: colorOK}
		if s.err != nil {
			status = cell{text: "invalid", color: colorError}
		}
		rows[i] = []cell{
			{text: s.name},
			{text: fmt.Sprintf("%d", s.actions)},
			{text: humanSize(s.size)},
			{text: s.modified.Format("2006-01-02 15:04")},
			status,
		}
	}
	cols := []column{{title: "NAME", expand: 2}, {title: "ACTIONS"}, {title: "SIZE"}, {title: "MODIFIED"}, {title: "STATUS", expand: 1}}
	arrow := " ▲"
	if v.sortDesc {
		arrow = " ▼"
	}
	cols[v.sortCol].title += arrow
	v.table.set(cols, rows)
	v.table.selectRow(0)
}

func (v *scriptsView) selected() (scriptInfo, bool) {
	i, ok := v.table.selected()
	if !ok {
		return scriptInfo{}, false
	}
	return v.filtered[i], true
}

// load reads a script into the scenario view and runs onLoaded on success.
func (v *scriptsView) load(info scriptInfo, onLoaded func()) {
	if open := v.app.scenario.script; open != nil && open.path == info.path {
		onLoaded()
		return
	}
	v.discardingScenario(func() {
		s, err := loadScript(info.path)
		if err != nil {
			v.app.flashErr(err)
			return
		}
		v.app.scenario.setScript(s)
		onLoaded()
	})
}

// discardingScenario runs fn, first confirming when the open scenario has unsaved changes fn would replace.
func (v *scriptsView) discardingScenario(fn func()) {
	if v.app.scenario.dirty() {
		v.app.confirm("Unsaved scenario changes will be lost, discard?", fn)
		return
	}
	fn()
}

// pickAction asks whether to open, execute or edit the selected script.
func (v *scriptsView) pickAction() {
	info, ok := v.selected()
	if !ok {
		return
	}
	v.app.pick(info.name, []string{"open", "execute", "edit raw script"}, func(choice string) {
		switch choice {
		case "open":
			v.load(info, func() { v.app.show(pageScenario) })
		case "execute":
			v.load(info, func() { v.app.scenario.execute() })
		default:
			v.editRaw(info)
		}
	})
}

// sortBy sorts by column col, reversing the order when already sorted by it.
func (v *scriptsView) sortBy(col int) {
	v.sortDesc = v.sortCol == col && !v.sortDesc
	v.sortCol = col
	v.render()
}

func (v *scriptsView) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		// the header row sits below the box border and the table x starts after it
		if !v.filtering && click.Button == tea.MouseLeft && click.Y == v.app.bodyY+1 {
			if col := v.table.columnAt(click.X - 1); col >= 0 && col < len(sortKeys) {
				v.sortBy(col)
			}
		}
		if !v.filtering && click.Button == tea.MouseLeft && click.Y > v.app.bodyY+1 {
			if row := v.table.rowAt(click.Y - v.app.bodyY - 1); row >= 0 {
				v.table.selectRow(row)
				v.pickAction()
			}
		}
		return
	}
	if delta := wheelDelta(msg); delta != 0 {
		v.table.scroll(delta)
		return
	}
	if v.filtering {
		v.filterInput(msg)
		return
	}

	key, ok := msg.(tea.KeyPressMsg)
	if !ok || v.table.key(key.String()) {
		return
	}
	if col := slices.Index(sortKeys, key.String()); col >= 0 {
		v.sortBy(col)
		return
	}
	switch key.String() {
	case "enter":
		if info, ok := v.selected(); ok {
			v.load(info, func() { v.app.show(pageScenario) })
		}
	case "/":
		v.filtering, v.filterOn = true, true
		v.filter.SetValue("")
		v.render()
	case "r":
		v.reload()
		v.app.flashInfo("script list reloaded")
	case "d":
		v.app.browseDir()
	case "x":
		if info, ok := v.selected(); ok {
			v.load(info, func() { v.app.scenario.execute() })
		}
	case "e":
		if info, ok := v.selected(); ok {
			v.editRaw(info)
		}
	}
}

func (v *scriptsView) filterInput(msg tea.Msg) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter":
			v.filtering = false
			return
		case "esc":
			v.filtering, v.filterOn = false, false
			v.filter.SetValue("")
			v.render()
			return
		case "up", "down":
			v.table.key(key.String())
			return
		}
	}
	v.filter, _ = v.filter.Update(msg)
	v.render()
}

func (v *scriptsView) view(w, h int) string {
	tableH := h
	filter := ""
	if v.filterOn {
		v.filter.SetWidth(w - 4)
		filter = "\n " + styleTitle.Render("/") + " " + v.filter.View()
		tableH--
	}
	content := v.table.view(w-2, tableH-2, true)
	if len(v.scripts) == 0 {
		banner := "No scripts in this folder\n\nPress " + styleKeyB.Render("<d>") + " to browse folders"
		content = center(banner, w-2, tableH-2)
	}
	return box(fmt.Sprintf(" Scripts [%d] ", len(v.filtered)), content, w, tableH, colorValue) + filter
}

// editRaw opens the whole script file in the JSON editor.
func (v *scriptsView) editRaw(info scriptInfo) {
	v.discardingScenario(func() { v.openRawEditor(info) })
}

func (v *scriptsView) openRawEditor(info scriptInfo) {
	s, err := loadScript(info.path)
	if err != nil {
		v.app.flashErr(err)
		return
	}
	v.app.editJSON(fmt.Sprintf("Edit %s", info.name), s.raw, func(edited []byte) error {
		saved := &script{path: s.path, raw: edited}
		actions, err := parseActions(edited)
		if err != nil {
			return err
		}
		saved.actions = actions
		if _, _, err := saved.config(); err != nil {
			return err
		}
		if err := saved.save(); err != nil {
			return err
		}
		v.app.scenario.setScript(saved)
		v.reload()
		v.app.flashInfo("saved %s", info.name)
		return nil
	})
}

func humanSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%dB", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	default:
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	}
}
