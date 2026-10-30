package tui

import (
	"fmt"
	"strings"
)

var helpSections = []struct {
	title string
	keys  [][2]string
}{
	{"General", [][2]string{
		{"?", "toggle this help"},
		{":", "command prompt"},
		{"esc", "back one level"},
		{"ctrl-c", "stop execution / quit"},
		{"/", "filter current list"},
		{"mouse", "click to select/edit, wheel to scroll, click breadcrumbs to navigate, click outside a dialog to close"},
	}},
	{"Scripts", [][2]string{
		{"enter", "open scenario"},
		{"x", "execute script"},
		{"e", "edit raw script JSON"},
		{"n/a/s/m", "sort by name/actions/size/modified, again reverses"},
		{"r", "reload script list"},
		{"d", "browse folders (popup)"},
	}},
	{"Scenario", [][2]string{
		{"tab", "switch between the Settings and Scenario tabs"},
		{"enter, e", "edit selected section or action in a form, enter on the actions field of an iterated action opens its sub-actions"},
		{"esc", "leave sub-actions, or go back to the script list"},
		{"ctrl-j", "edit selected section or action as JSON"},
		{"a", "add section / action after selection"},
		{"ctrl-d", "delete selected section or action"},
		{"d", "enable/disable action"},
		{"y", "duplicate action"},
		{"J, K", "move action down/up"},
		{"u", "undo last change"},
		{"c", "edit raw script JSON"},
		{"v", "validate script"},
		{"ctrl-s", "write script to disk"},
		{"x", "execute script"},
	}},
	{"Editor", [][2]string{
		{"ctrl-s", "validate and save"},
		{"ctrl-j", "switch form to JSON editing"},
		{"tab, ↓, ↑", "move between fields, opens dropdowns with enter"},
		{"enter", "on a compound field, edit it in a sub-form (or raw JSON for lists/maps)"},
		{"ctrl-u", "unset the focused field"},
		{"ctrl-f", "format JSON"},
		{"esc", "cancel"},
	}},
	{"Execution", [][2]string{
		{"ctrl-c", "stop execution"},
		{"x", "re-run script"},
		{"v", "switch between log and graph view"},
		{"f", "filter log by level"},
		{"F", "toggle log follow"},
		{"w", "toggle line wrap"},
		{"c", "clear log"},
		{"g, G", "jump to top/bottom"},
	}},
	{"Commands", [][2]string{
		{":scripts", "script browser"},
		{":scenario", "scenario of current script"},
		{":execute", "execute current script"},
		{":logs", "execution view"},
		{":reload", "reload script list"},
		{":dir", "change working directory"},
		{":quit", "quit"},
	}},
}

func (a *App) showHelp() {
	buf := &strings.Builder{}
	for _, section := range helpSections {
		_, _ = fmt.Fprintf(buf, "%s\n", styleTitleB.Render(section.title))
		for _, key := range section.keys {
			_, _ = fmt.Fprintf(buf, "  %s %s\n", styleKeyB.Render(fmt.Sprintf("%-12s", key[0])), key[1])
		}
		_, _ = fmt.Fprintln(buf)
	}

	a.showStyledText("Help", buf.String())
}
