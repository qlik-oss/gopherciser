package tui

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/statistics"
)

const (
	execLogBuffer = 4096
	execMaxLines  = 20000
	execTrimSlack = 5000
	statsHeight   = 5
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

type (
	// logEntry is a log message as rendered in the execution view.
	logEntry struct {
		time         time.Time
		level        logger.LogLevel
		action       string
		label        string
		message      string
		details      string
		infoType     string
		responseTime int64
		success      bool
		raw          bool // written to stdout/stderr rather than through the logger
	}

	// tuiWriter is a gopherciser log writer feeding the execution view.
	tuiWriter struct {
		entries chan<- logEntry
	}

	// execView runs a script and shows its log output live.
	execView struct {
		app   *App
		logs  viewport.Model
		lines []string // rendered log lines currently shown
		graph *graphView

		mu       sync.Mutex
		script   *script
		counters *statistics.ExecutionCounters
		started  time.Time
		stopped  time.Time
		result   string
		active   atomic.Bool
		cancel   context.CancelFunc

		entries   []logEntry // everything but errors and warnings
		problems  []logEntry // errors and warnings, kept apart so chatty levels can't push them out
		follow    bool
		wrap      bool
		filters   []logFilter
		showGraph bool
	}

	// logFilter toggles the display of one log level, or of console output when raw.
	logFilter struct {
		label  string
		level  logger.LogLevel
		raw    bool
		hidden bool
	}

	// logMsg carries a batch of log entries to the UI.
	logMsg struct{ entries []logEntry }
	// statsMsg asks the UI to sample the execution counters.
	statsMsg struct{}
	// execDoneMsg reports that the script finished.
	execDoneMsg struct{ err error }
)

// WriteMessage implements logger.MsgWriter.
func (w *tuiWriter) WriteMessage(msg *logger.LogChanMsg) error {
	if msg == nil {
		return nil
	}
	entry := logEntry{
		time:         msg.Time,
		level:        msg.Level,
		action:       msg.Action,
		label:        msg.Label,
		message:      msg.Message,
		details:      msg.Details,
		infoType:     msg.InfoType,
		responseTime: msg.ResponseTime,
		success:      msg.Success,
	}
	select {
	case w.entries <- entry:
	default: // never block the executing script on a slow UI
	}
	return nil
}

// Level implements logger.MsgWriter.
func (w *tuiWriter) Level(logger.LogLevel) {}

func newExecView(app *App) *execView {
	v := &execView{
		app:       app,
		logs:      viewport.New(),
		graph:     newGraphView(),
		follow:    true,
		showGraph: true,
		filters: []logFilter{
			{label: "result", level: logger.ResultLevel},
			{label: "error", level: logger.ErrorLevel},
			{label: "warning", level: logger.WarningLevel},
			{label: "info", level: logger.InfoLevel},
			{label: "metrics", level: logger.MetricsLevel},
			{label: "traffic", level: logger.TrafficLevel},
			{label: "debug", level: logger.DebugLevel},
			{label: "console output", raw: true},
		},
	}
	v.logs.MouseWheelEnabled = false
	return v
}

func (v *execView) activate() {
	hints := [][2]string{
		{"ctrl-c", "stop execution"},
		{"x", "re-run script"},
		{"v", "log/graph view"},
	}
	if !v.showGraph {
		hints = append(hints, [2]string{"f", "filter log"}, [2]string{"F", "toggle follow"})
	}
	v.app.setHints(append(hints,
		[2]string{"w", "toggle wrap"},
		[2]string{"c", "clear log"},
		[2]string{"g/G", "top/bottom"},
		[2]string{"esc", "back"},
		[2]string{":", "command"},
		[2]string{"?", "help"},
	))
}

func (v *execView) running() bool { return v.active.Load() }

func (v *execView) capturing() bool { return false }

func (v *execView) update(msg tea.Msg) {
	if delta := wheelDelta(msg); delta != 0 {
		v.scroll(delta)
		return
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}

	onOff := map[bool]string{true: "on", false: "off"}
	switch key.String() {
	case "v":
		v.showGraph = !v.showGraph
		v.activate()
	case "F":
		if v.showGraph {
			return
		}
		v.follow = !v.follow
		v.app.flashInfo("follow %s", onOff[v.follow])
	case "w":
		v.wrap = !v.wrap
		v.logs.SoftWrap = v.wrap
		v.app.flashInfo("wrap %s", onOff[v.wrap])
	case "f":
		if v.showGraph {
			return
		}
		v.app.openOverlay(&filterDialog{app: v.app, filters: v.filters, onChange: v.renderLogs})
	case "c":
		v.entries, v.problems = nil, nil
		v.renderLogs()
	case "x":
		v.rerun()
	default:
		if !v.showGraph {
			scrollKey(&v.logs, key.String())
		}
	}
}

func (v *execView) scroll(delta int) {
	if v.showGraph {
		return
	}
	if delta < 0 {
		v.logs.ScrollUp(-delta)
	} else {
		v.logs.ScrollDown(delta)
	}
}

func (v *execView) view(w, h int) string {
	stats := box(" Execution ", v.statsText(), w, statsHeight, colorValue)
	bodyH := h - statsHeight
	if v.showGraph {
		return stats + "\n" + box(" Graphs ", v.graph.view(w-2, bodyH-2), w, bodyH, colorValue)
	}

	v.logs.SetWidth(w - 2)
	v.logs.SetHeight(bodyH - 2)
	if v.follow {
		v.logs.GotoBottom()
	}
	return stats + "\n" + box(v.logTitle(), v.logs.View(), w, bodyH, colorValue)
}

func (v *execView) rerun() {
	if v.running() {
		v.app.flashWarn("script is already running")
		return
	}
	v.mu.Lock()
	s := v.script
	v.mu.Unlock()
	if s == nil {
		return
	}
	cfg, _, err := s.config()
	if err != nil {
		v.app.flashErr(err)
		return
	}
	v.start(s, cfg)
}

// start executes cfg, streaming log entries into the view.
func (v *execView) start(s *script, cfg *config.Config) {
	if v.running() {
		v.app.flashWarn("script is already running")
		return
	}

	v.mu.Lock()
	v.script = s
	v.counters = &cfg.Counters
	v.started = time.Now()
	v.stopped = time.Time{}
	v.result = "running"
	v.mu.Unlock()

	v.entries, v.problems = nil, nil
	v.follow = true
	v.renderLogs()
	v.graph.reset()
	v.active.Store(true)

	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	entries := make(chan logEntry, execLogBuffer)
	cfg.CustomLoggers = append(cfg.CustomLoggers, logger.NewLogger(&tuiWriter{entries: entries}))
	var entriesMu sync.Mutex
	entriesClosed := false
	cfg.Cancel = func(msg string) {
		entriesMu.Lock()
		if !entriesClosed {
			select {
			case entries <- logEntry{time: time.Now(), level: logger.ErrorLevel, message: msg}:
			default:
			}
		}
		entriesMu.Unlock()
		cancel()
	}

	// sampling follows the execution itself, not the cancel context, since a
	// stopped run keeps producing counters until all threads have finished
	done := make(chan struct{})
	go v.consume(entries)
	go v.pollStats(done)

	go func() {
		var err error
		defer cancel()
		defer func() {
			entriesMu.Lock()
			entriesClosed = true
			close(entries)
			entriesMu.Unlock()
		}()
		defer func() { v.finish(err) }()
		defer close(done)

		// gopherciser writes status, summary and console logs to stdout/stderr,
		// capture those to keep them from corrupting the screen.
		defer captureOutput(entries)()

		templateData := struct{ ConfigFile string }{
			strings.TrimSuffix(filepath.Base(s.path), filepath.Ext(s.path)),
		}
		err = cfg.Execute(ctx, templateData)
	}()
}

func (v *execView) stop() {
	if !v.running() {
		return
	}
	v.app.flashWarn("stopping execution...")
	if v.cancel != nil {
		v.cancel()
	}
}

// captureOutput redirects stdout and stderr into the log view, the returned
// function restores them.
func captureOutput(entries chan<- logEntry) func() {
	reader, writer, err := os.Pipe()
	if err != nil {
		return func() {}
	}
	restore := redirectStdio(writer)

	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		scanner.Split(scanConsoleLines)
		for scanner.Scan() {
			line := scanner.Text()
			plain := strings.TrimSpace(ansiPattern.ReplaceAllString(line, ""))
			// periodic status line, already shown in the execution pane
			if plain == "" || strings.Contains(plain, "ActvSess<") {
				continue
			}
			select {
			case entries <- logEntry{time: time.Now(), level: logger.InfoLevel, message: line, raw: true}:
			default:
			}
		}
	}()

	return func() {
		restore()
		_ = writer.Close()
		<-done
		_ = reader.Close()
	}
}

// scanConsoleLines splits on both newline and carriage return, so status lines
// rewritten in place show up as separate entries.
func scanConsoleLines(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (v *execView) finish(err error) {
	v.mu.Lock()
	v.stopped = time.Now()
	switch {
	case err != nil:
		v.result = truncate("failed: "+oneLine(err.Error()), 60)
	default:
		v.result = "completed"
	}
	v.mu.Unlock()

	v.active.Store(false)
	v.app.post(execDoneMsg{err: err})
}

// finished reports the outcome of the execution, must run on the UI goroutine.
func (v *execView) finished(err error) {
	if err != nil {
		v.app.flashErr(err)
		return
	}
	v.mu.Lock()
	result := v.result
	v.mu.Unlock()
	v.app.flashInfo("execution %s", result)
}

// consume forwards log entries to the UI in batches to keep drawing cheap.
func (v *execView) consume(entries <-chan logEntry) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	pending := make([]logEntry, 0, 256)
	flush := func() {
		if len(pending) < 1 {
			return
		}
		batch := pending
		pending = make([]logEntry, 0, 256)
		v.app.post(logMsg{entries: batch})
	}

	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				flush()
				return
			}
			pending = append(pending, entry)
			if len(pending) >= 256 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// append stores and shows a batch of entries, must run on the UI goroutine.
func (v *execView) append(batch []logEntry) {
	trimmed := false
	for _, entry := range batch {
		if isProblem(entry) {
			v.problems = append(v.problems, entry)
		} else {
			v.entries = append(v.entries, entry)
		}
		if v.visible(entry) {
			v.lines = append(v.lines, formatEntry(entry, true))
		}
	}
	// trim in chunks so the log is not rebuilt on every batch
	if len(v.entries) > execMaxLines+execTrimSlack {
		v.entries = v.entries[len(v.entries)-execMaxLines:]
		trimmed = true
	}
	if len(v.problems) > execMaxLines+execTrimSlack {
		v.problems = v.problems[len(v.problems)-execMaxLines:]
		trimmed = true
	}
	if trimmed {
		v.renderLogs()
		return
	}
	v.logs.SetContentLines(v.lines)
}

func isProblem(entry logEntry) bool {
	return !entry.raw && (entry.level == logger.ErrorLevel || entry.level == logger.WarningLevel)
}

func (v *execView) visible(entry logEntry) bool {
	for _, f := range v.filters {
		if f.hidden && f.raw == entry.raw && (entry.raw || f.level == entry.level) {
			return false
		}
	}
	return true
}

func (v *execView) filtered() bool {
	for _, f := range v.filters {
		if f.hidden {
			return true
		}
	}
	return false
}

// renderLogs redraws the log pane from the stored entries.
func (v *execView) renderLogs() {
	v.lines = v.lines[:0]
	rest, problems := v.entries, v.problems
	for len(rest) > 0 || len(problems) > 0 {
		var entry logEntry
		if len(problems) == 0 || len(rest) > 0 && !problems[0].time.Before(rest[0].time) {
			entry, rest = rest[0], rest[1:]
		} else {
			entry, problems = problems[0], problems[1:]
		}
		if v.visible(entry) {
			v.lines = append(v.lines, formatEntry(entry, true))
		}
	}
	v.logs.SetContentLines(v.lines)
	if v.follow {
		v.logs.GotoBottom()
	}
}

func (v *execView) logTitle() string {
	v.mu.Lock()
	s := v.script
	v.mu.Unlock()

	title := " Log"
	if s != nil {
		title += ": " + filepath.Base(s.path)
	}
	title += fmt.Sprintf(" [%d]", len(v.entries)+len(v.problems))
	if v.filtered() {
		title += " (filtered)"
	}
	return title + " "
}

// pollStats asks the UI to sample the counters once a second.
func (v *execView) pollStats(done <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	v.app.post(statsMsg{})
	for {
		select {
		case <-done:
			v.app.post(statsMsg{})
			return
		case <-ticker.C:
			v.app.post(statsMsg{})
		}
	}
}

// sample records the counters for the graphs, must run on the UI goroutine.
func (v *execView) sample() {
	v.mu.Lock()
	counters := v.counters
	v.mu.Unlock()
	if counters != nil {
		v.graph.record(counters, time.Now())
	}
}

func (v *execView) statsText() string {
	v.mu.Lock()
	counters, started, stopped, result, s := v.counters, v.started, v.stopped, v.result, v.script
	v.mu.Unlock()

	if counters == nil || s == nil {
		return " no execution started, select a script and press " + styleKeyB.Render("<x>")
	}

	elapsed := time.Since(started).Truncate(time.Second)
	if !stopped.IsZero() {
		elapsed = stopped.Sub(started).Truncate(time.Second)
	}

	state := fg(colorOK).Bold(true)
	switch {
	case v.running():
		state = fg(colorTitle).Bold(true)
	case strings.HasPrefix(result, "failed"):
		state = fg(colorError).Bold(true)
	}

	errs, warns := counters.Errors.Current(), counters.Warnings.Current()
	errStyle, warnStyle := fg(colorValue).Bold(true), fg(colorValue).Bold(true)
	if errs > 0 {
		errStyle = fg(colorError).Bold(true)
	}
	if warns > 0 {
		warnStyle = fg(colorWarning).Bold(true)
	}

	key := func(name string) string { return styleKey.Render(name + ":") }
	return fmt.Sprintf(" %s %s   %s %s   %s %s\n", key("Script"), filepath.Base(s.path), key("State"), state.Render(result), key("Elapsed"), elapsed) +
		fmt.Sprintf(" %s %d   %s %d   %s %d   %s %d   %s %d\n",
			key("Threads"), counters.Threads.Current(), key("Sessions"), counters.Sessions.Current(),
			key("Active users"), counters.ActiveUsers.Current(), key("Actions"), counters.ActionID.Current(),
			key("Requests"), counters.Requests.Current()) +
		fmt.Sprintf(" %s %s   %s %s", key("Errors"), errStyle.Render(fmt.Sprint(errs)), key("Warnings"), warnStyle.Render(fmt.Sprint(warns)))
}

// sgrOnly strips everything but colors from terminal escape sequences, so
// captured console output can not move the cursor or otherwise corrupt the UI.
func sgrOnly(text string) string {
	text = ansiPattern.ReplaceAllStringFunc(text, func(seq string) string {
		if strings.HasSuffix(seq, "m") {
			return seq
		}
		return ""
	})
	return strings.Map(func(r rune) rune {
		if r == '\t' {
			return ' '
		}
		if r < 0x20 && r != 0x1b || r == 0x7f {
			return -1
		}
		return r
	}, text)
}

// formatEntry renders an entry as one line, with colors when colored.
func formatEntry(entry logEntry, colored bool) string {
	timestamp := entry.time.Format("15:04:05")
	if entry.raw {
		if !colored {
			return fmt.Sprintf("%s %s", timestamp, ansi.Strip(entry.message))
		}
		return fmt.Sprintf("%s %s\x1b[0m", styleDisabled.Render(timestamp), sgrOnly(entry.message))
	}

	style := fg(colorValue)
	switch entry.level {
	case logger.ErrorLevel:
		style = fg(colorError)
	case logger.WarningLevel:
		style = fg(colorWarning)
	case logger.ResultLevel:
		style = fg(colorOK)
	case logger.DebugLevel, logger.TrafficLevel, logger.MetricsLevel:
		style = styleDisabled
	}

	action := entry.action
	if entry.label != "" {
		action = fmt.Sprintf("%s/%s", entry.action, entry.label)
	}
	if action != "" {
		action = fmt.Sprintf(" [%s]", action)
	}

	message := entry.message
	if entry.infoType != "" {
		message = strings.TrimSpace(fmt.Sprintf("%s %s", strings.ToLower(entry.infoType), message))
	}
	if message == "" {
		message = entry.details
	} else if entry.details != "" {
		message = fmt.Sprintf("%s: %s", message, entry.details)
	}
	if entry.level == logger.ResultLevel {
		state := "success"
		if !entry.success {
			state = "failed"
		}
		message = strings.TrimPrefix(fmt.Sprintf("%s %s %dms", message, state, entry.responseTime), " ")
	}
	message = oneLine(message)

	if !colored {
		return fmt.Sprintf("%s %-7s%s %s", timestamp, entry.level.String(), action, message)
	}
	if action != "" {
		action = " " + styleTitleB.Render(action[1:])
	}
	return fmt.Sprintf("%s %s%s %s", styleDisabled.Render(timestamp), style.Render(fmt.Sprintf("%-7s", entry.level.String())), action, message)
}
