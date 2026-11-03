package tui

import (
	"bytes"
	"fmt"
	"image/color"
	"reflect"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/buger/jsonparser"
)

const (
	// unsetOption is the dropdown entry removing the property from the JSON.
	unsetOption = "(unset)"

	formHelpHeight  = 6
	dropdownVisible = 8
	textAreaRows    = 5
)

// form is a modal editing a JSON object through generated fields.
type form struct {
	app    *App
	spec   *formSpec
	title  string
	raw    []byte
	onSave func([]byte) error

	current []byte
	fields  []*formField
	focus   int // a field, then the Save and Cancel buttons
	scroll  int

	// hit data of the last draw
	width      int
	visible    int   // body lines shown
	lineItems  []int // field index per layout line, -1 when none
	buttonLine int
	saveX      [2]int // button ranges inside the border
	cancelX    [2]int
	optLine    int // layout line of the first dropdown option shown
	optFirst   int // option index shown on optLine

	status      string
	statusColor color.Color
}

// editForm opens a generated form for raw. onSave receives the resulting JSON
// and its error, if any, is shown without closing the form.
func (a *App) editForm(spec *formSpec, title string, raw []byte, onSave func([]byte) error) {
	f := &form{app: a, spec: spec, title: title, raw: raw, current: raw, onSave: onSave}
	f.build(0)
	a.openOverlay(f)
}

func (f *form) floating() bool { return false }

func (f *form) setStatus(c color.Color, format string, args ...any) {
	f.status, f.statusColor = fmt.Sprintf(format, args...), c
}

// build regenerates the fields from the current JSON.
func (f *form) build(focus int) {
	f.fields = f.spec.fields(f.current)
	for _, field := range f.fields {
		field.initWidget()
	}
	f.setFocus(focus)
}

func (f *form) saveButton() int   { return len(f.fields) }
func (f *form) cancelButton() int { return len(f.fields) + 1 }

func (f *form) setFocus(i int) {
	f.focus = min(max(i, 0), f.cancelButton())
	for j, field := range f.fields {
		field.setFocused(j == f.focus)
	}
}

// focused returns the focused field, nil when a button has focus.
func (f *form) focused() *formField {
	if f.focus < len(f.fields) {
		return f.fields[f.focus]
	}
	return nil
}

func (f *form) changed() bool {
	for _, field := range f.fields {
		if field.changed() {
			return true
		}
	}
	return !bytes.Equal(f.current, f.raw)
}

func (f *form) merged() ([]byte, error) { return applyFields(f.current, f.fields) }

func (f *form) close() { f.app.closeOverlay(f) }

func (f *form) cancel() {
	if f.changed() {
		f.app.confirm("Discard changes?", f.close)
		return
	}
	f.close()
}

func (f *form) save() {
	edited, err := f.merged()
	if err == nil {
		err = f.spec.validate(edited)
	}
	if err == nil {
		err = f.onSave(edited)
	}
	if err != nil {
		f.setStatus(colorError, "%v", err)
		return
	}
	f.close()
}

func (f *form) editAsJSON() {
	edited, err := f.merged()
	if err != nil {
		f.setStatus(colorError, "%v", err)
		return
	}
	f.close()
	f.app.editJSON(f.title, edited, f.onSave)
}

// drill opens a compound field in an editor of its own.
func (f *form) drill(field *formField) {
	if f.spec.openActions != nil && samePath(field.path, []string{"settings", "actions"}) {
		f.spec.openActions()
		f.save()
		return
	}
	display := field.display
	if display == "" {
		display = field.name()
	}
	onFieldSave := func(edited []byte) error {
		field.value = compact(edited)
		field.set = true
		field.sync()
		return nil
	}
	if field.elemType != nil {
		raw := []byte(field.value)
		if len(raw) < 1 {
			raw = []byte("{}")
		}
		f.app.editForm(&formSpec{root: field.elemType}, display, raw, onFieldSave)
		return
	}
	f.app.editJSON(display, []byte(field.value), onFieldSave)
}

// switchType rebuilds the form with the properties of the newly selected type.
func (f *form) switchType(typeField *formField) {
	if typeField.value == typeField.orig {
		return
	}
	updated, err := applyFields(f.current, f.fields)
	if err != nil {
		f.setStatus(colorError, "%v", err)
		return
	}
	if _, path := f.spec.selector.resolve(typeField.value); len(path) > 0 {
		updated = jsonparser.Delete(updated, path...)
	}
	f.current = updated
	f.build(0)
}

func (f *form) click(click tea.MouseClickMsg) {
	y, x := click.Y-f.app.bodyY-1, click.X-1
	if click.Button != tea.MouseLeft || y < 0 || y >= f.visible || x < 0 || x >= f.width-2 {
		return
	}
	line := y + f.scroll
	if line >= len(f.lineItems) {
		return
	}
	if line == f.buttonLine {
		switch {
		case x >= f.saveX[0] && x < f.saveX[1]:
			f.save()
		case x >= f.cancelX[0] && x < f.cancelX[1]:
			f.cancel()
		}
		return
	}
	i := f.lineItems[line]
	if i < 0 {
		return
	}

	field := f.fields[i]
	if field.open {
		switch {
		case line == f.optLine-1:
			field.open = false
		case line >= f.optLine:
			field.cursor = f.optFirst + line - f.optLine
			if field.handleDropdown("enter", true) && f.spec.selector != nil && f.focus == 0 {
				f.switchType(field)
			}
		}
		return
	}
	wasFocused := i == f.focus
	f.setFocus(i)
	switch {
	case field.kind == kindBool:
		field.handle(tea.KeyPressMsg{Code: tea.KeySpace})
	case field.kind == kindEnum:
		field.open = true
	case field.compound && wasFocused:
		f.drill(field)
	}
}

// wheel moves the focus, or the cursor of an open dropdown.
func (f *form) wheel(delta int) {
	step := 1
	if delta < 0 {
		step = -1
	}
	if field := f.focused(); field != nil && field.open {
		field.cursor = min(max(field.cursor+step, 0), len(field.opts)-1)
		return
	}
	f.setFocus(f.focus + step)
}

func (f *form) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		f.click(click)
		return
	}
	if delta := wheelDelta(msg); delta != 0 {
		f.wheel(delta)
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		if paste, ok := msg.(tea.PasteMsg); ok {
			if field := f.focused(); field != nil {
				field.handle(paste)
			}
		}
		return
	}

	field := f.focused()
	switch key.String() {
	case "ctrl+s":
		f.save()
		return
	case "ctrl+u":
		if field != nil {
			field.unset()
		}
		return
	case "ctrl+j":
		f.editAsJSON()
		return
	case "esc":
		if field != nil && field.open {
			field.open = false
			return
		}
		f.cancel()
		return
	case "tab":
		f.setFocus(f.focus + 1)
		return
	case "shift+tab":
		f.setFocus(f.focus - 1)
		return
	case "up", "down":
		// arrows move between fields, except where the item uses them itself
		if field == nil || !field.usesVertical() {
			if key.String() == "down" {
				f.setFocus(f.focus + 1)
			} else {
				f.setFocus(f.focus - 1)
			}
			return
		}
	case "enter":
		switch {
		case field == nil:
			if f.focus == f.saveButton() {
				f.save()
			} else {
				f.cancel()
			}
			return
		case field.compound:
			f.drill(field)
			return
		case field.kind == kindText || field.kind == kindNumber || field.kind == kindJSON:
			f.setFocus(f.focus + 1)
			return
		}
	case "left", "right":
		if field == nil {
			if f.focus == f.saveButton() {
				f.setFocus(f.cancelButton())
			} else {
				f.setFocus(f.saveButton())
			}
			return
		}
	}

	if field == nil {
		return
	}
	if field.handle(msg) && field.kind == kindEnum && !field.open && f.spec.selector != nil && f.focus == 0 {
		f.switchType(field)
	}
}

// layout renders all fields and buttons, returning the lines and the first
// line of the focused item.
func (f *form) layout(w int) (lines []string, focusStart, focusEnd int) {
	labelWidth := 0
	for _, field := range f.fields {
		labelWidth = max(labelWidth, lipgloss.Width(field.name()))
	}
	labelWidth = min(labelWidth, w/2)
	widgetWidth := max(w-labelWidth-3, 10)

	f.lineItems = f.lineItems[:0]
	for i, field := range f.fields {
		label := fit(field.name(), labelWidth)
		if i == f.focus {
			label = styleTitleB.Render(label)
		} else {
			label = styleKey.Render(label)
		}
		block := lipgloss.JoinHorizontal(lipgloss.Top, " "+label+" ", field.render(widgetWidth, i == f.focus))
		if i == f.focus {
			focusStart = len(lines)
			focusEnd = focusStart + lipgloss.Height(block)
		}
		if field.open {
			f.optLine, f.optFirst = len(lines)+1, field.firstOption()
		}
		for range lipgloss.Height(block) {
			f.lineItems = append(f.lineItems, i)
		}
		lines = append(lines, strings.Split(block, "\n")...)
	}

	button := func(label string, i int) string {
		style := lipgloss.NewStyle().Padding(0, 1).Background(colorKey).Foreground(colorValue)
		if f.focus == i {
			style = styleSelected.Padding(0, 1)
		}
		return style.Render(label)
	}
	save, cancel := button("Save", f.saveButton()), button("Cancel", f.cancelButton())
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, save, " ", cancel)
	if f.focus >= len(f.fields) {
		focusStart = len(lines) + 1
		focusEnd = focusStart + 1
	}
	x := w - 1 - lipgloss.Width(buttons)
	f.saveX = [2]int{x, x + lipgloss.Width(save)}
	f.cancelX = [2]int{f.saveX[1] + 1, x + lipgloss.Width(buttons)}
	f.buttonLine = len(lines) + 1
	f.lineItems = append(f.lineItems, -1, -1)
	lines = append(lines, "", lipgloss.PlaceHorizontal(w-1, lipgloss.Right, buttons))
	return lines, focusStart, focusEnd
}

func (f *form) helpText() string {
	field := f.focused()
	if field == nil {
		return ""
	}
	display := field.display
	if display == "" {
		display = field.path[len(field.path)-1]
	}
	state := "unset, property omitted"
	if field.set {
		state = "set"
	}

	var help strings.Builder
	help.WriteString(styleTitleB.Render(display) + " " + styleDisabled.Render(fmt.Sprintf("(%s, %s)", field.name(), state)) + "\n")
	docs := field.docs()
	if len(docs) < 1 {
		help.WriteString(styleDisabled.Render("no documentation"))
	}
	for _, line := range docs {
		help.WriteString(line + "\n")
	}
	return help.String()
}

func (f *form) view(w, h int) string {
	bodyH := h - formHelpHeight - 1
	lines, start, end := f.layout(w - 2)

	visible := bodyH - 2
	f.width, f.visible = w, visible
	if end > f.scroll+visible {
		f.scroll = end - visible
	}
	if start < f.scroll {
		f.scroll = start
	}
	f.scroll = min(max(f.scroll, 0), max(len(lines)-visible, 0))
	body := strings.Join(lines[f.scroll:], "\n")

	help := lipgloss.NewStyle().Width(w - 2).Render(f.helpText())

	status := " " + strings.Join([]string{
		hint("ctrl-s", "save"), hint("ctrl-u", "unset field"), hint("enter", "open sub-form/JSON"),
		hint("ctrl-j", "edit as JSON"), hint("tab/↓/↑", "move"), hint("esc", "cancel"),
	}, "  ")
	if f.status != "" {
		status = fg(f.statusColor).Render(" " + oneLine(f.status))
	}

	return strings.Join([]string{
		box(" "+f.title+" ", body, w, bodyH, colorValue),
		box(" Help ", help, w, formHelpHeight, colorValue),
		status,
	}, "\n")
}

// initWidget creates the widget editing the field.
func (f *formField) initWidget() {
	switch f.kind {
	case kindEnum:
		f.opts = f.options
		if f.value != "" && !contains(f.opts, f.value) {
			f.opts = append([]string{f.value}, f.opts...)
		}
		f.opts = append([]string{unsetOption}, f.opts...)
		f.open = false
		f.cursor = indexOf(f.opts, optionLabel(f.value))
	case kindTextArea:
		f.area = textarea.New()
		f.area.ShowLineNumbers = false
		f.area.MaxHeight = 0
		f.area.MaxWidth = 0
		f.area.SetVirtualCursor(true)
		f.area.SetPromptFunc(0, func(textarea.PromptInfo) string { return "" })
		f.area.SetHeight(textAreaRows)
		f.area.SetValue(f.value)
		f.area.MoveToBegin()
		f.area.Blur()
	case kindBool:
	default:
		f.input = newInput(f.value)
		f.input.Blur()
	}
	f.sync()
}

func (f *formField) setFocused(focused bool) {
	switch f.kind {
	case kindEnum:
		f.open = f.open && focused
	case kindTextArea:
		if focused {
			f.area.Focus()
		} else {
			f.area.Blur()
		}
	case kindBool:
	default:
		if focused {
			f.input.Focus()
		} else {
			f.input.Blur()
		}
	}
}

// usesVertical reports whether the widget uses the up and down keys itself.
func (f *formField) usesVertical() bool {
	return f.kind == kindTextArea || (f.kind == kindEnum && f.open)
}

// sync updates the widget to reflect the value and set state.
func (f *formField) sync() {
	switch f.kind {
	case kindEnum:
		f.cursor = indexOf(f.opts, optionLabel(f.value))
	case kindTextArea:
		if f.area.Value() != f.value {
			f.area.SetValue(f.value)
		}
	case kindBool:
	default:
		switch {
		case !f.set:
			f.input.Placeholder = unsetOption
		case f.goKind == reflect.String && f.kind == kindText:
			f.input.Placeholder = `""` // set but empty
		default:
			f.input.Placeholder = ""
		}
		if f.input.Value() != f.value {
			f.input.SetValue(f.value)
			f.input.CursorEnd()
		}
	}
}

// handle applies input to the widget, reporting whether the value was changed.
func (f *formField) handle(msg tea.Msg) bool {
	key, isKey := msg.(tea.KeyPressMsg)
	switch f.kind {
	case kindBool:
		if isKey && (key.String() == "space" || key.String() == "enter") {
			f.value = fmt.Sprintf("%t", f.value != "true")
			f.set = true
			return true
		}
	case kindEnum:
		return f.handleDropdown(key.String(), isKey)
	case kindTextArea:
		f.area, _ = f.area.Update(msg)
		if text := f.area.Value(); text != f.value {
			f.value, f.set = text, true
			return true
		}
	default:
		before := f.input.Value()
		f.input, _ = f.input.Update(msg)
		text := f.input.Value()
		if f.kind == kindNumber && strings.Trim(text, "+-.0123456789eE") != "" {
			f.input.SetValue(before)
			f.input.CursorEnd()
			return false
		}
		if text != f.value {
			f.value, f.set = text, true
			f.sync()
			return true
		}
	}
	return false
}

func (f *formField) handleDropdown(key string, isKey bool) bool {
	if !isKey {
		return false
	}
	if !f.open {
		if key == "enter" || key == "space" {
			f.open = true
		}
		return false
	}
	switch key {
	case "up":
		f.cursor = max(f.cursor-1, 0)
	case "down":
		f.cursor = min(f.cursor+1, len(f.opts)-1)
	case "enter", "space":
		f.open = false
		f.value = optionValue(f.opts[f.cursor])
		f.set = f.value != ""
		return true
	}
	return false
}

// firstOption is the first dropdown entry shown in the open list.
func (f *formField) firstOption() int {
	return min(max(f.cursor-dropdownVisible+1, 0), max(len(f.opts)-dropdownVisible, 0))
}

// render draws the widget in width cells.
func (f *formField) render(width int, focused bool) string {
	switch f.kind {
	case kindBool:
		box := "[ ]"
		switch {
		case !f.set:
			box = "[-]" // unset, distinct from false
		case f.value == "true":
			box = "[*]"
		}
		return box
	case kindEnum:
		current := optionLabel(f.value)
		if !f.open {
			return fit(current+" ▼", width)
		}
		first := f.firstOption()
		lines := []string{fit(current+" ▲", width)}
		for i := first; i < len(f.opts) && i < first+dropdownVisible; i++ {
			line := fit(" "+f.opts[i], width)
			if i == f.cursor {
				line = styleSelected.Render(line)
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	case kindTextArea:
		f.area.SetWidth(width)
		view := f.area.View()
		if !f.set {
			view = styleDisabled.Render(unsetOption) + "\n" + view
		}
		return view
	default:
		f.input.SetWidth(width)
		return f.input.View()
	}
}

func optionLabel(value string) string {
	if value == "" {
		return unsetOption
	}
	return value
}

func optionValue(label string) string {
	if label == unsetOption {
		return ""
	}
	return label
}

func contains(list []string, s string) bool { return indexOf(list, s) >= 0 }

func indexOf(list []string, s string) int {
	for i, item := range list {
		if item == s {
			return i
		}
	}
	return -1
}
