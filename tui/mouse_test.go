package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	a := NewApp(t.TempDir())
	a.w, a.h = 120, 40
	return a
}

// find returns the screen cell of the first occurrence of text.
func find(t *testing.T, a *App, text string) (x, y int) {
	t.Helper()
	for y, line := range strings.Split(ansi.Strip(a.render()), "\n") {
		if i := strings.Index(line, text); i >= 0 {
			return ansi.StringWidth(line[:i]), y
		}
	}
	t.Fatalf("%q not on screen", text)
	return 0, 0
}

func click(a *App, x, y int) { a.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}) }

func TestMouseConfirm(t *testing.T) {
	for _, tc := range []struct {
		name     string
		target   string // text to click, empty for outside the dialog
		wantOK   bool
		wantOpen bool
	}{
		{"ok", " OK ", true, false},
		{"cancel", " Cancel ", false, false},
		{"outside", "", false, false},
		{"text", "sure?", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp(t)
			ran := false
			a.confirm("sure?", func() { ran = true })
			x, y := 0, a.h-3
			if tc.target != "" {
				x, y = find(t, a, tc.target)
				x++
			}
			click(a, x, y)
			if ran != tc.wantOK || (len(a.overlays) > 0) != tc.wantOpen {
				t.Fatalf("ran=%v open=%v", ran, len(a.overlays) > 0)
			}
		})
	}
}

func TestMouseBreadcrumb(t *testing.T) {
	a := newTestApp(t)
	a.show(pageExec)
	x, y := find(t, a, "scripts >")
	click(a, x, y)
	if a.page != pageScripts {
		t.Fatalf("page %q", a.page)
	}
}

func TestMouseForm(t *testing.T) {
	a := newTestApp(t)
	saved := false
	a.editForm(sectionSpec("connectionSettings"), "t", []byte(`{"server":"localhost"}`), func([]byte) error {
		saved = true
		return nil
	})
	f := a.top().(*form)
	a.render()
	second := 0
	for f.lineItems[second] != 1 {
		second++
	}
	click(a, 5, a.bodyY+1+second)
	if f.focus != 1 {
		t.Fatalf("focus %d", f.focus)
	}
	x, y := find(t, a, "Save")
	click(a, x+1, y)
	if !saved || len(a.overlays) != 0 {
		t.Fatalf("saved=%v open=%d", saved, len(a.overlays))
	}
}

func TestMouseScenarioRows(t *testing.T) {
	a := newTestApp(t)
	s, err := loadScript(writeTestScript(t, testScript))
	if err != nil {
		t.Fatal(err)
	}
	a.scenario.setScript(s)
	a.show(pageScenario)
	x, y := find(t, a, "openapp")
	click(a, x, y)
	if a.scenario.actions.cursor != 1 || len(a.overlays) != 0 {
		t.Fatalf("cursor %d overlays %d", a.scenario.actions.cursor, len(a.overlays))
	}
	a.render()
	click(a, x, y)
	if len(a.overlays) != 1 {
		t.Fatal("second click should open the editor")
	}
}
