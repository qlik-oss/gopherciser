package tui

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/goccy/go-json"
	"github.com/qlik-oss/gopherciser/scenario"
)

const (
	maxHistory      = 100
	maxSectionRows  = 6
	settingsPreview = 90
)

// scenarioView edits a single script: its top level sections and the actions
// of the scenario array.
type scenarioView struct {
	app      *App
	sections table
	actions  table
	// inSections is set when the sections table has focus
	inSections bool
	// tabRanges are the x ranges of the Settings and Scenario tabs, inside the border
	tabRanges [2][2]int

	script *script
	// path holds the indexes of the parent actions down to the shown actions, empty at the top level
	path []int
	// cur is the list of actions shown, the scenario or the sub actions of the action at path
	cur     []action
	history []*script // states before each change, for undo
	savedAt int       // history length matching the file on disk
}

func newScenarioView(app *App) *scenarioView {
	return &scenarioView{app: app}
}

// setScript replaces the edited script, dropping any unsaved state.
func (v *scenarioView) setScript(s *script) {
	v.script = s
	v.history = nil
	v.savedAt = 0
	v.path = nil
	v.refresh()
}

// refresh reloads the shown actions from the script, going up as far as needed when path no longer exists.
func (v *scenarioView) refresh() {
	for {
		list := v.script.actions
		ok := true
		for _, i := range v.path {
			if i >= len(list) || !list[i].hasSubActions() {
				ok = false
				break
			}
			sub, err := list[i].subActions()
			if err != nil {
				ok = false
				break
			}
			list = sub
		}
		if ok {
			v.cur = list
			return
		}
		v.path = v.path[:len(v.path)-1]
	}
}

// commit writes the shown actions back into the script.
func (v *scenarioView) commit() {
	root, err := replaceActionsAt(v.script.actions, v.path, v.cur)
	if err != nil {
		v.app.flashErr(err)
		return
	}
	v.script.actions = root
}

// open shows the sub actions of the selected action, reporting if there were any to show.
func (v *scenarioView) open() bool {
	i, ok := v.index()
	if !ok || !v.cur[i].hasSubActions() {
		return false
	}
	sub, err := v.cur[i].subActions()
	if err != nil {
		v.app.flashErr(err)
		return true
	}
	v.path = append(v.path, i)
	v.cur = sub
	v.render()
	v.actions.selectRow(0)
	return true
}

// up goes one level towards the scenario, reporting if there was a level to leave.
func (v *scenarioView) up() bool {
	if v.script == nil || len(v.path) == 0 {
		return false
	}
	i := v.path[len(v.path)-1]
	v.path = v.path[:len(v.path)-1]
	v.refresh()
	v.render()
	v.actions.selectRow(i)
	return true
}

// level describes where the shown actions are, e.g. "Scenario" or "Scenario > 3 iterated".
func (v *scenarioView) level() string {
	name := "Scenario"
	list := v.script.actions
	for _, i := range v.path {
		name += fmt.Sprintf(" > %d %s", i+1, list[i].Type)
		list, _ = list[i].subActions()
	}
	return name
}

func (v *scenarioView) dirty() bool { return len(v.history) != v.savedAt }

func (v *scenarioView) capturing() bool { return false }

func (v *scenarioView) activate() {
	v.setTab(false)
	v.render()
}

// setTab shows the Settings tab when inSections, else the Scenario tab, with
// the key hints of that tab.
func (v *scenarioView) setTab(inSections bool) {
	v.inSections = inSections
	hints := [][2]string{
		{"enter/e", "edit"},
		{"ctrl-j", "edit as JSON"},
		{"a", "add"},
		{"ctrl-d", "delete"},
		{"tab", "settings/scenario"},
	}
	if !inSections {
		hints = append(hints,
			[2]string{"d", "enable/disable"},
			[2]string{"y", "duplicate action"},
			[2]string{"J/K", "move down/up"})
	}
	v.app.setHints(append(hints,
		[2]string{"u", "undo"},
		[2]string{"ctrl-s", "save"},
		[2]string{"x", "execute"},
		[2]string{"v", "validate"},
		[2]string{"c", "edit raw script"},
		[2]string{"esc", "back"},
		[2]string{"?", "help"}))
}

// snapshot records the current state so the next change can be undone.
func (v *scenarioView) snapshot() {
	v.history = append(v.history, v.script.clone())
	if len(v.history) > maxHistory {
		v.history = v.history[1:]
		if v.savedAt > 0 {
			v.savedAt--
		}
	}
}

func (v *scenarioView) undo() {
	if len(v.history) < 1 {
		v.app.flashWarn("nothing to undo")
		return
	}
	v.undoSilently()
	v.render()
	v.app.flashInfo("undone, %d change(s) left", len(v.history))
}

func (v *scenarioView) render() {
	if v.script == nil {
		v.sections.set(nil, nil)
		v.actions.set(nil, nil)
		return
	}

	sections, err := v.script.sections()
	if err != nil {
		v.app.flashErr(err)
	}
	rows := make([][]cell, len(sections))
	for i, sec := range sections {
		rows[i] = []cell{{text: sec.name, color: colorKey}, {text: truncate(compact(sec.raw), settingsPreview), color: colorDisabled}}
	}
	v.sections.set([]column{{title: "SECTION", max: 24}, {title: "VALUE", expand: 1}}, rows)

	rows = make([][]cell, len(v.cur))
	for i, act := range v.cur {
		state := cell{text: "enabled", color: colorOK}
		if act.Disabled {
			state = cell{text: "disabled", color: colorDisabled}
		}
		rows[i] = []cell{
			{text: fmt.Sprintf("%d", i+1)},
			{text: act.Label},
			{text: act.Type},
			state,
			{text: truncate(act.settingsOf(), settingsPreview), color: colorDisabled},
		}
	}
	v.actions.set([]column{{title: "#", max: 4}, {title: "LABEL", expand: 1}, {title: "ACTION", expand: 1}, {title: "STATE"}, {title: "SETTINGS", expand: 3}}, rows)
}

func (v *scenarioView) view(w, h int) string {
	if v.script == nil {
		return box(" Scenario ", "", w, h, colorValue)
	}

	name := filepath.Base(v.script.path)
	if v.dirty() {
		name += "*"
	}
	tabs := []string{"Settings", fmt.Sprintf("%s [%d]", v.level(), len(v.cur))}
	var row strings.Builder
	row.WriteString(" ")
	pos := 1
	for i, name := range tabs {
		label := " " + name + " "
		style := styleHeader
		if (i == 0) == v.inSections {
			style = styleSelected
		}
		row.WriteString(style.Render(label) + " ")
		v.tabRanges[i] = [2]int{pos, pos + len(label)}
		pos += len(label) + 1
	}
	body := v.actions.view(w-2, h-3, !v.inSections)
	if v.inSections {
		body = v.sections.view(w-2, h-3, true)
	}
	return box(" "+name+" ", row.String()+"\n"+body, w, h, colorValue)
}

func (v *scenarioView) active() *table {
	if v.inSections {
		return &v.sections
	}
	return &v.actions
}

// index returns the selected action index.
func (v *scenarioView) index() (int, bool) {
	if v.script == nil {
		return 0, false
	}
	return v.actions.selected()
}

// selectedSection returns the selected section, if any.
func (v *scenarioView) selectedSection() (int, section, bool) {
	if v.script == nil {
		return 0, section{}, false
	}
	sections, err := v.script.sections()
	if err != nil {
		return 0, section{}, false
	}
	i, ok := v.sections.selected()
	if !ok || i >= len(sections) {
		return 0, section{}, false
	}
	return i, sections[i], true
}

func (v *scenarioView) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		// the tab row sits below the box border
		if click.Button == tea.MouseLeft && click.Y == v.app.bodyY+1 {
			for i, r := range v.tabRanges {
				if x := click.X - 1; x >= r[0] && x < r[1] {
					v.setTab(i == 0)
				}
			}
		}
		// the table header sits below the tab row, rows start under it
		if click.Button == tea.MouseLeft && click.Y > v.app.bodyY+2 {
			if row := v.active().rowAt(click.Y - v.app.bodyY - 2); row >= 0 {
				if cur, _ := v.active().selected(); cur == row {
					v.editSelected(false)
				} else {
					v.active().selectRow(row)
				}
			}
		}
		return
	}
	if delta := wheelDelta(msg); delta != 0 {
		v.active().scroll(delta)
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok || v.active().key(key.String()) {
		return
	}

	switch key.String() {
	case "tab", "shift+tab":
		v.setTab(!v.inSections)
	case "ctrl+d":
		if v.inSections {
			v.deleteSection()
		} else {
			v.deleteAction()
		}
	case "ctrl+j":
		v.editSelected(true)
	case "ctrl+s":
		v.save()
	case "enter", "e":
		v.editSelected(false)
	case "a":
		if v.inSections {
			v.addSection()
		} else {
			v.addAction()
		}
	case "d":
		v.toggleDisabled()
	case "y":
		v.duplicateAction()
	case "J":
		v.move(1)
	case "K":
		v.move(-1)
	case "u":
		v.undo()
	case "x":
		v.execute()
	case "v":
		v.validate()
	case "c":
		v.editRaw()
	}
}

// editSelected edits the selected section or action, as JSON when asJSON.
func (v *scenarioView) editSelected(asJSON bool) {
	if v.inSections {
		v.editSection(asJSON)
		return
	}
	v.editAction(asJSON)
}

func (v *scenarioView) editSection(asJSON bool) {
	i, sec, ok := v.selectedSection()
	if !ok {
		return
	}
	v.openEditor(sectionSpec(sec.name), "Section: "+sec.name, sec.raw, asJSON, func(edited []byte) error {
		if bytes.Equal(compactBytes(edited), compactBytes(sec.raw)) {
			return nil
		}
		if err := v.updateSection(sec.name, edited); err != nil {
			return err
		}
		v.render()
		v.sections.selectRow(i)
		v.app.flashInfo("%s updated, <ctrl-s> to save", sec.name)
		return nil
	})
}

// openEditor edits raw with a generated form when spec is known and asJSON is
// false, otherwise with the JSON editor.
func (v *scenarioView) openEditor(spec *formSpec, title string, raw []byte, asJSON bool, onSave func([]byte) error) {
	saveToDisk := onSave
	onSave = func(edited []byte) error {
		if err := saveToDisk(edited); err != nil {
			return err
		}
		v.save()
		return nil
	}
	if spec == nil || asJSON {
		v.app.editJSON(title, raw, onSave)
		return
	}
	v.app.editForm(spec, title, raw, onSave)
}

// updateSection replaces a section, keeping the change undoable and rejecting
// it when the script no longer validates.
func (v *scenarioView) updateSection(name string, raw []byte) error {
	v.snapshot()
	if err := v.script.setSection(name, raw); err != nil {
		v.undoSilently()
		return err
	}
	if _, _, err := v.script.config(); err != nil {
		v.undoSilently()
		return err
	}
	return nil
}

func (v *scenarioView) undoSilently() {
	v.script = v.history[len(v.history)-1]
	v.history = v.history[:len(v.history)-1]
	v.refresh()
}

func (v *scenarioView) addSection() {
	if v.script == nil {
		return
	}
	existing := map[string]bool{}
	if sections, err := v.script.sections(); err == nil {
		for _, sec := range sections {
			existing[sec.name] = true
		}
	}
	var missing []string
	for _, name := range knownSections {
		if !existing[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) < 1 {
		v.app.flashWarn("all known sections already exist, edit them with <e>")
		return
	}
	v.app.pick("Add section", missing, func(name string) {
		v.openEditor(sectionSpec(name), "Section: "+name, []byte("{}"), false, func(edited []byte) error {
			if err := v.updateSection(name, edited); err != nil {
				return err
			}
			v.render()
			v.app.flashInfo("%s added, <ctrl-s> to save", name)
			return nil
		})
	})
}

func (v *scenarioView) deleteSection() {
	_, sec, ok := v.selectedSection()
	if !ok {
		return
	}
	v.app.confirm(fmt.Sprintf("Delete section %q?", sec.name), func() {
		v.snapshot()
		v.script.deleteSection(sec.name)
		v.render()
		v.app.flashInfo("%s deleted, <ctrl-s> to save", sec.name)
	})
}

func (v *scenarioView) editAction(asJSON bool) {
	i, ok := v.index()
	if !ok {
		return
	}
	act := v.cur[i]
	title := fmt.Sprintf("Action %d: %s", i+1, act.Type)

	spec := actionSpec()
	openSub := false
	spec.openActions = func() { openSub = true }
	v.openEditor(spec, title, act.raw, asJSON, func(edited []byte) error {
		wantSub := openSub
		openSub = false
		if bytes.Equal(compactBytes(edited), compactBytes(act.raw)) {
			if wantSub {
				v.open()
			}
			return nil
		}
		updated := act
		if err := updated.setRaw(edited); err != nil {
			return err
		}
		v.snapshot()
		v.cur[i] = updated
		v.commit()
		v.render()
		v.actions.selectRow(i)
		v.app.flashInfo("action %d updated, <ctrl-s> to save", i+1)
		if wantSub {
			v.open()
		}
		return nil
	})
}

// insertAction adds act after the selected action and opens it for editing.
func (v *scenarioView) insertAction(act action, edit bool) {
	insertAt := len(v.cur)
	if i, ok := v.index(); ok {
		insertAt = i + 1
	}
	v.snapshot()
	v.cur = append(v.cur, action{})
	copy(v.cur[insertAt+1:], v.cur[insertAt:])
	v.cur[insertAt] = act
	v.commit()
	v.render()
	v.actions.selectRow(insertAt)
	if edit {
		v.editAction(false)
	}
}

func (v *scenarioView) addAction() {
	if v.script == nil {
		return
	}
	types := scenario.RegisteredActions()
	sort.Strings(types)
	v.app.pick("Add action", types, func(actionType string) {
		v.insertAction(newAction(actionType), true)
	})
}

func (v *scenarioView) duplicateAction() {
	i, ok := v.index()
	if !ok {
		return
	}
	act := v.cur[i]
	act.raw = bytes.Clone(act.raw)
	v.insertAction(act, false)
	v.app.flashInfo("action %d duplicated, <ctrl-s> to save", i+1)
}

func (v *scenarioView) deleteAction() {
	i, ok := v.index()
	if !ok {
		return
	}
	act := v.cur[i]
	v.app.confirm(fmt.Sprintf("Delete action %d (%s)?", i+1, act.Type), func() {
		v.snapshot()
		v.cur = append(v.cur[:i], v.cur[i+1:]...)
		v.commit()
		v.render()
		v.app.flashInfo("action deleted, <ctrl-s> to save")
	})
}

func (v *scenarioView) toggleDisabled() {
	i, ok := v.index()
	if !ok {
		return
	}
	act := v.cur[i]
	raw, err := setBool(act.raw, !act.Disabled, "disabled")
	if err != nil {
		v.app.flashErr(err)
		return
	}
	if err := act.setRaw(raw); err != nil {
		v.app.flashErr(err)
		return
	}
	v.snapshot()
	v.cur[i] = act
	v.commit()
	v.render()
	v.actions.selectRow(i)
	v.app.flashInfo("action %d %s, <ctrl-s> to save", i+1, map[bool]string{true: "disabled", false: "enabled"}[act.Disabled])
}

func (v *scenarioView) move(offset int) {
	i, ok := v.index()
	if !ok {
		return
	}
	target := i + offset
	if target < 0 || target >= len(v.cur) {
		return
	}
	v.snapshot()
	v.cur[i], v.cur[target] = v.cur[target], v.cur[i]
	v.commit()
	v.render()
	v.actions.selectRow(target)
}

func (v *scenarioView) save() {
	if v.script == nil {
		return
	}
	if _, warnings, err := v.script.config(); err != nil {
		v.app.flashErr(err)
		return
	} else if len(warnings) > 0 {
		v.app.flashWarn("saved with %d validation warning(s), <v> for details", len(warnings))
	}
	if err := v.script.save(); err != nil {
		v.app.flashErr(err)
		return
	}
	v.savedAt = len(v.history)
	v.render()
	if len(v.script.actions) > 0 {
		v.app.flashInfo("saved %s", filepath.Base(v.script.path))
	}
}

func (v *scenarioView) validate() {
	if v.script == nil {
		return
	}
	_, warnings, err := v.script.config()
	if err != nil {
		v.app.flashErr(err)
		return
	}
	if len(warnings) < 1 {
		v.app.flashInfo("script is valid")
		return
	}
	text := fmt.Sprintf("Script is valid with %d warning(s):\n", len(warnings))
	for i, warning := range warnings {
		text += fmt.Sprintf("\n%d. %s", i+1, warning)
	}
	v.app.showText("Validation warnings", text)
}

// editRaw opens the full script document in the editor.
func (v *scenarioView) editRaw() {
	if v.script == nil {
		return
	}
	raw, err := v.script.document()
	if err != nil {
		v.app.flashErr(err)
		return
	}
	v.app.editJSON(fmt.Sprintf("Edit %s", filepath.Base(v.script.path)), raw, func(edited []byte) error {
		if bytes.Equal(compactBytes(edited), compactBytes(raw)) {
			return nil
		}
		updated := &script{path: v.script.path, raw: edited}
		actions, err := parseActions(edited)
		if err != nil {
			return err
		}
		updated.actions = actions
		if _, _, err := updated.config(); err != nil {
			return err
		}
		v.snapshot()
		v.script = updated
		v.refresh()
		v.render()
		v.app.flashInfo("script updated, <ctrl-s> to save")
		return nil
	})
}

func (v *scenarioView) execute() {
	if v.script == nil {
		return
	}
	cfg, warnings, err := v.script.config()
	if err != nil {
		v.app.flashErr(err)
		return
	}
	if len(warnings) > 0 {
		v.app.flashWarn("%d validation warning(s), <v> for details", len(warnings))
	}

	v.app.exec.start(v.script, cfg)
	v.app.show(pageExec)
}

// compact returns raw JSON without insignificant whitespace.
func compact(raw []byte) string { return string(compactBytes(raw)) }

func compactBytes(raw []byte) []byte {
	buf := bytes.NewBuffer(nil)
	if err := json.Compact(buf, raw); err != nil {
		return raw
	}
	return buf.Bytes()
}
