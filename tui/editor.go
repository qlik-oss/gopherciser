package tui

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/goccy/go-json"
)

// jsonEditor is a modal JSON text editor.
type jsonEditor struct {
	app    *App
	title  string
	area   textarea.Model
	orig   string
	onSave func([]byte) error

	status      string
	statusColor color.Color
}

// editJSON opens a modal JSON editor. onSave is called with the edited content
// and its error, if any, is shown without closing the editor.
func (a *App) editJSON(title string, content []byte, onSave func([]byte) error) {
	pretty := content
	if formatted, err := formatJSON(content); err == nil {
		pretty = formatted
	}

	area := textarea.New()
	area.ShowLineNumbers = false
	area.MaxHeight = 0
	area.MaxWidth = 0
	area.SetVirtualCursor(true)
	area.SetPromptFunc(0, func(textarea.PromptInfo) string { return "" })
	area.SetValue(string(pretty))
	area.MoveToBegin()
	area.Focus()

	a.openOverlay(&jsonEditor{app: a, title: title, area: area, orig: area.Value(), onSave: onSave})
}

func (e *jsonEditor) floating() bool { return false }

func (e *jsonEditor) setStatus(c color.Color, format string, args ...any) {
	e.status, e.statusColor = fmt.Sprintf(format, args...), c
}

func (e *jsonEditor) update(msg tea.Msg) {
	if delta := wheelDelta(msg); delta != 0 {
		for ; delta < 0; delta++ {
			e.area.CursorUp()
		}
		for ; delta > 0; delta-- {
			e.area.CursorDown()
		}
		return
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "ctrl+s":
			e.save()
			return
		case "ctrl+f":
			formatted, err := formatJSON([]byte(e.area.Value()))
			if err != nil {
				e.setStatus(colorError, "%v", err)
				return
			}
			e.area.SetValue(string(formatted))
			e.setStatus(colorOK, "formatted")
			return
		case "esc":
			if e.area.Value() == e.orig {
				e.app.closeOverlay(e)
				return
			}
			e.app.confirm("Discard changes?", func() { e.app.closeOverlay(e) })
			return
		}
	}
	e.area, _ = e.area.Update(msg)
}

func (e *jsonEditor) save() {
	edited := []byte(e.area.Value())
	if !json.Valid(edited) {
		e.setStatus(colorError, "invalid JSON")
		return
	}
	if err := e.onSave(edited); err != nil {
		e.setStatus(colorError, "%v", err)
		return
	}
	e.app.closeOverlay(e)
}

func (e *jsonEditor) view(w, h int) string {
	e.area.SetWidth(w - 2)
	e.area.SetHeight(h - 3)

	status := " " + strings.Join([]string{hint("ctrl-s", "save"), hint("ctrl-f", "format"), hint("esc", "cancel")}, "  ")
	if e.status != "" {
		status = fg(e.statusColor).Render(" " + oneLine(e.status))
	}
	return box(" "+e.title+" ", e.area.View(), w, h-1, colorValue) + "\n" + status
}

func formatJSON(raw []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.Indent(buf, raw, "", indent); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return buf.Bytes(), nil
}
