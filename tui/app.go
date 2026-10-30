// Package tui implements a terminal UI for browsing, editing and
// executing gopherciser scenario scripts.
package tui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	pageScripts  = "scripts"
	pageScenario = "scenario"
	pageExec     = "exec"

	// header layout: fixed info column, key hints laid out in as many columns as fit
	infoWidth    = 52
	hintColWidth = 27
	minHintRows  = 3
	maxHintRows  = 5

	flashDuration = 5 * time.Second
)

type (
	// App is the gopherciser terminal UI, a Bubble Tea model.
	App struct {
		dir string

		startOnScenario bool
		w, h            int
		send            func(tea.Msg)
		cmds            []tea.Cmd

		page     string
		scripts  *scriptsView
		scenario *scenarioView
		exec     *execView
		overlays []overlay

		prompt   textinput.Model
		promptOn bool

		hintList    [][2]string
		crumbRanges []crumbRange
		bodyY       int // screen row where the page body starts
		overlayX    int // screen position of the floating overlay
		overlayY    int
		flash       string
		flashColor  color.Color
		flashID     int
	}

	// screen is a page of the app.
	screen interface {
		activate()
		update(msg tea.Msg)
		view(w, h int) string
		// capturing reports whether the screen is reading text, keeping
		// the global keys from being triggered.
		capturing() bool
	}

	// overlay is a modal on top of the current screen, receiving all input.
	overlay interface {
		update(msg tea.Msg)
		view(w, h int) string
		// floating overlays are drawn as a small box over the screen, others
		// take over the whole body.
		floating() bool
	}

	// crumbRange is the clickable x range of a breadcrumb on the status line.
	crumbRange struct {
		x0, x1 int
		page   string
	}

	flashExpiredMsg struct{ id int }
)

// NewApp creates the TUI rooted at dir.
func NewApp(dir string) *App {
	app := &App{dir: dir}
	app.prompt = newInput("")
	app.scripts = newScriptsView(app)
	app.scenario = newScenarioView(app)
	app.exec = newExecView(app)
	return app
}

// OpenScript loads the script at path into the scenario view on start.
func (a *App) OpenScript(path string) error {
	s, err := loadScript(path)
	if err != nil {
		return err
	}
	a.scenario.setScript(s)
	a.startOnScenario = true
	return nil
}

// Run starts the UI.
func (a *App) Run() error {
	out, closeOut := ttyOutput()
	defer closeOut()

	prog := tea.NewProgram(a, tea.WithOutput(out))
	a.send = prog.Send
	a.show(pageScripts)
	if a.startOnScenario {
		a.show(pageScenario)
	}
	_, err := prog.Run()
	return err
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd { return a.drain() }

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
	case tea.KeyPressMsg:
		a.key(msg)
	case tea.PasteMsg, tea.MouseWheelMsg, tea.MouseClickMsg:
		a.input(msg)
	case flashExpiredMsg:
		if msg.id == a.flashID {
			a.flash = ""
		}
	case logMsg:
		a.exec.append(msg.entries)
	case statsMsg:
		a.exec.sample()
	case execDoneMsg:
		a.exec.finished(msg.err)
	}
	return a, a.drain()
}

// View implements tea.Model.
func (a *App) View() tea.View {
	v := tea.NewView(a.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// post sends a message to the running program from any goroutine.
func (a *App) post(msg tea.Msg) {
	if a.send != nil {
		a.send(msg)
	}
}

// queue schedules a command to be returned from the current update.
func (a *App) queue(cmd tea.Cmd) { a.cmds = append(a.cmds, cmd) }

func (a *App) drain() tea.Cmd {
	cmd := tea.Batch(a.cmds...)
	a.cmds = nil
	return cmd
}

func (a *App) screen() screen {
	switch a.page {
	case pageScenario:
		return a.scenario
	case pageExec:
		return a.exec
	default:
		return a.scripts
	}
}

func (a *App) top() overlay {
	if n := len(a.overlays); n > 0 {
		return a.overlays[n-1]
	}
	return nil
}

func (a *App) openOverlay(o overlay) { a.overlays = append(a.overlays, o) }

func (a *App) closeOverlay(o overlay) {
	for i, existing := range a.overlays {
		if existing == o {
			a.overlays = append(a.overlays[:i], a.overlays[i+1:]...)
			return
		}
	}
}

// input routes non key input to whatever currently receives it.
func (a *App) input(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok && a.top() == nil && !a.promptOn &&
		click.Button == tea.MouseLeft && click.Y == a.h-1 {
		for _, r := range a.crumbRanges {
			if click.X >= r.x0 && click.X < r.x1 && r.page != a.page {
				a.show(r.page)
			}
		}
		return
	}
	switch {
	case a.top() != nil:
		a.top().update(msg)
	case a.promptOn:
		a.prompt, _ = a.prompt.Update(msg)
	default:
		a.screen().update(msg)
	}
}

func (a *App) key(msg tea.KeyPressMsg) {
	if o := a.top(); o != nil {
		o.update(msg)
		return
	}
	if a.promptOn {
		a.promptKey(msg)
		return
	}
	if a.screen().capturing() {
		a.screen().update(msg)
		return
	}

	switch msg.String() {
	case "ctrl+c":
		if a.page == pageExec && a.exec.running() {
			a.exec.stop()
			return
		}
		a.quit()
	case "esc":
		a.back()
	case "?":
		a.showHelp()
	case ":":
		a.showPrompt()
	default:
		a.screen().update(msg)
	}
}

// show switches to a page and lets it refresh itself.
func (a *App) show(page string) {
	a.page = page
	a.screen().activate()
}

// back navigates one level up, exec view first asking to stop a running script.
func (a *App) back() {
	switch a.page {
	case pageScenario:
		if a.scenario.up() {
			return
		}
		if a.scenario.dirty() {
			a.confirm("Unsaved changes will be lost, discard?", func() { a.show(pageScripts) })
			return
		}
		a.show(pageScripts)
	case pageExec:
		if a.exec.running() {
			a.flashWarn("script is running, ctrl-c to stop")
			return
		}
		if a.scenario.script != nil {
			a.show(pageScenario)
			return
		}
		a.show(pageScripts)
	default:
		a.quit()
	}
}

func (a *App) quit() {
	if a.exec.running() {
		a.confirm("A script is running, quit anyway?", func() {
			a.exec.stop()
			a.queue(tea.Quit)
		})
		return
	}
	a.queue(tea.Quit)
}

func (a *App) runCommand(cmd string) {
	switch strings.ToLower(cmd) {
	case "":
	case "q", "quit", "exit":
		a.quit()
	case "s", "scripts":
		a.show(pageScripts)
	case "sc", "scenario":
		if a.scenario.script == nil {
			a.flashWarn("no script selected")
			return
		}
		a.show(pageScenario)
	case "x", "exec", "execute":
		if a.scenario.script == nil {
			a.flashWarn("no script selected")
			return
		}
		a.scenario.execute()
	case "log", "logs":
		a.show(pageExec)
	case "h", "help", "?":
		a.showHelp()
	case "r", "reload":
		a.show(pageScripts)
		a.flashInfo("script list reloaded")
	case "d", "dir", "cd":
		a.browseDir()
	case "path":
		a.changeDir()
	default:
		a.flashErr(fmt.Errorf("unknown command: %s", cmd))
	}
}

func (a *App) showPrompt() {
	a.prompt.SetValue("")
	a.prompt.Focus()
	a.promptOn = true
}

func (a *App) promptKey(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "enter":
		cmd := strings.TrimSpace(a.prompt.Value())
		a.promptOn = false
		a.runCommand(cmd)
	case "esc":
		a.promptOn = false
	default:
		a.prompt, _ = a.prompt.Update(msg)
	}
}

// setHints sets the key hints shown in the header.
func (a *App) setHints(hints [][2]string) { a.hintList = hints }

// layoutHints renders the hints column by column in the given width, growing
// the header when they need more rows and dropping trailing hints that do not
// fit.
func (a *App) layoutHints(width int) []string {
	cols := max(width/hintColWidth, 1)
	rows := min(max((len(a.hintList)+cols-1)/cols, minHintRows), maxHintRows)
	shown := min(len(a.hintList), rows*cols)
	cols = max((shown+rows-1)/rows, 1)

	lines := make([]string, rows)
	for row := range lines {
		for col := 0; col < cols; col++ {
			i := col*rows + row
			if i >= shown {
				continue
			}
			cell := hint(a.hintList[i][0], a.hintList[i][1])
			if col < cols-1 {
				cell = fit(cell, hintColWidth)
			}
			lines[row] += cell
		}
	}
	return lines
}

func (a *App) header() string {
	dir := a.dir
	if max := infoWidth - 7; len(dir) > max {
		dir = "…" + dir[len(dir)-max+1:]
	}
	info := []string{
		" " + styleTitleB.Render("Gopherciser") + " TUI",
		" " + styleKey.Render("Dir:") + " " + dir,
	}

	hints := a.layoutHints(a.w - infoWidth - 2)
	lines := make([]string, len(hints))
	for i, h := range hints {
		left := ""
		if i < len(info) {
			left = info[i]
		}
		lines[i] = fit(left, infoWidth) + h
	}
	return box("", strings.Join(lines, "\n"), a.w, len(lines)+2, colorHeader)
}

// crumbs renders the breadcrumbs starting at screen column x.
func (a *App) crumbs(x int) string {
	crumbs := [][2]string{{"scripts", pageScripts}}
	switch a.page {
	case pageScenario:
		crumbs = append(crumbs, [2]string{"scenario", pageScenario})
	case pageExec:
		if a.scenario.script != nil {
			crumbs = append(crumbs, [2]string{"scenario", pageScenario})
		}
		crumbs = append(crumbs, [2]string{"execute", pageExec})
	}

	a.crumbRanges = a.crumbRanges[:0]
	names := make([]string, len(crumbs))
	x++ // leading space
	for i, c := range crumbs {
		names[i] = c[0]
		a.crumbRanges = append(a.crumbRanges, crumbRange{x, x + len(c[0]), c[1]})
		x += len(c[0]) + len(" > ")
	}
	return styleReverse.Render(" " + strings.Join(names, " > ") + " ")
}

func (a *App) statusLine() string {
	left := a.w * 3 / 5
	flash := ""
	if a.flash != "" {
		flash = fg(a.flashColor).Render(" " + oneLine(a.flash))
	}
	return fit(flash, left) + fit(a.crumbs(left), a.w-left)
}

func (a *App) render() string {
	if a.w < 40 || a.h < 12 {
		return "terminal too small"
	}

	header := a.header()
	lines := []string{header}
	a.bodyY = lipgloss.Height(header)
	bodyH := a.h - a.bodyY - 1
	if a.promptOn {
		bodyH--
	}

	body := a.screen().view(a.w, bodyH)
	if o := a.top(); o != nil {
		if o.floating() {
			dialog := o.view(a.w, bodyH)
			a.overlayX = (a.w - lipgloss.Width(dialog)) / 2
			a.overlayY = a.bodyY + (bodyH-lipgloss.Height(dialog))/2
			body = lipgloss.NewCompositor(
				lipgloss.NewLayer(body),
				lipgloss.NewLayer(dialog).
					X((a.w-lipgloss.Width(dialog))/2).
					Y((bodyH-lipgloss.Height(dialog))/2).
					Z(1),
			).Render()
		} else {
			body = o.view(a.w, bodyH)
		}
	}
	lines = append(lines, body)

	if a.promptOn {
		a.prompt.SetWidth(a.w - 3)
		lines = append(lines, styleTitle.Render(":")+" "+a.prompt.View())
	}
	lines = append(lines, a.statusLine())
	return strings.Join(lines, "\n")
}

func (a *App) flashInfo(format string, args ...any) { a.setFlash(colorOK, format, args...) }
func (a *App) flashWarn(format string, args ...any) { a.setFlash(colorWarning, format, args...) }
func (a *App) flashErr(err error)                   { a.setFlash(colorError, "%v", err) }

func (a *App) setFlash(c color.Color, format string, args ...any) {
	a.flashID++
	id := a.flashID
	a.flash = fmt.Sprintf(format, args...)
	a.flashColor = c
	a.queue(tea.Tick(flashDuration, func(time.Time) tea.Msg { return flashExpiredMsg{id} }))
}

// newInput creates a single line text input holding value.
func newInput(value string) textinput.Model {
	in := textinput.New()
	in.Prompt = ""
	in.SetVirtualCursor(true)
	in.SetValue(value)
	in.CursorEnd()
	in.Focus()
	return in
}

// changeDir prompts for a new working directory and, once it validates,
// switches to it and reloads the script list.
// browseDir shows a popup of the folders below the working directory. Picking
// a folder opens it, ".." goes up and "." keeps the current folder.
func (a *App) browseDir() {
	items := []string{". (use this folder)", ".."}
	if entries, err := os.ReadDir(a.dir); err == nil {
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				items = append(items, e.Name())
			}
		}
	}
	a.pick("Folder: "+a.dir, items, func(name string) {
		if strings.HasPrefix(name, ".") && name != ".." {
			return
		}
		a.dir = filepath.Join(a.dir, name)
		a.scripts.reload()
		a.browseDir()
	})
}

func (a *App) changeDir() {
	d := &dirDialog{app: a, input: newInput(a.dir)}
	a.openOverlay(d)
}

type dirDialog struct {
	app    *App
	input  textinput.Model
	width  int
	height int
}

func (d *dirDialog) floating() bool { return true }

func (d *dirDialog) update(msg tea.Msg) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		x, y := click.X-d.app.overlayX, click.Y-d.app.overlayY
		if x < 0 || x >= d.width || y < 0 || y >= d.height {
			d.app.closeOverlay(d)
			return
		}
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			d.app.closeOverlay(d)
			return
		case "enter":
			d.apply()
			return
		}
	}
	d.input, _ = d.input.Update(msg)
}

func (d *dirDialog) apply() {
	a := d.app
	dir := strings.TrimSpace(d.input.Value())
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			a.flashErr(fmt.Errorf("change dir: %w", err))
			return
		}
		dir = filepath.Join(home, dir[1:])
	}
	info, err := os.Stat(dir)
	if err != nil {
		a.flashErr(fmt.Errorf("change dir: %w", err))
		return
	}
	if !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	a.closeOverlay(d)
	a.dir = dir
	a.scripts.reload()
	a.flashInfo("working directory set to %s", dir)
}

func (d *dirDialog) view(w, _ int) string {
	width, height := min(60, w-4), 3
	d.width, d.height = width, height
	d.input.SetWidth(width - 8)
	return box(" Change working directory ", " "+styleTitle.Render("Dir:")+" "+d.input.View(), width, height, colorValue)
}

// oneLine collapses a multi line message into a single line.
func oneLine(text string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(text, "\n", " ")), " ")
}

// wheelDelta returns how many rows a mouse wheel event scrolls, negative up.
func wheelDelta(msg tea.Msg) int {
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		switch wheel.Button {
		case tea.MouseWheelUp:
			return -3
		case tea.MouseWheelDown:
			return 3
		}
	}
	return 0
}
