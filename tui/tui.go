package tui

import (
	"context"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/statistics"
)

type (
	Values struct {
		// Threads - Total started threads
		Threads uint64
		// Sessions - Total started sessions
		Sessions uint64
		// Users - Total unique users
		Users uint64
		// Warnings - Total warnings
		Warnings uint64
		// Requests - Total requests sentq
		Requests uint64
		// ActiveUsers - Currently active users
		ActiveUsers uint64
		// AppCounter -  App counter for round robin access
		// AppCounter int
		// Errors
		Errors uint64
	}

	Progress struct {
		Values        map[string][]float64
		pushedCounter int
	}
)

const (
	// MaxX = 100

	Threads     = "threads"
	Sessions    = "sessions"
	Users       = "users"
	Warnings    = "warnings"
	Requests    = "requests"
	ActiveUsers = "activeusers"
	Errors      = "errors"
)

func StartProgressTui(ctx context.Context, cancel func(), counters *statistics.ExecutionCounters) error {
	if counters == nil {
		return errors.Errorf("no statistics counters provided")
	}

	progress := NewProgress()
	startTextBox := widgets.NewParagraph()
	startTextBox.Title = "Gopherciser"
	startTextBox.Text = "Load test starting up..."

	startTextBox.SetRect(5, 5, 50, 25)
	startTextBox.BorderStyle.Fg = termui.ColorWhite
	termui.Render(startTextBox)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				progress.pushValues(Values{
					Threads:     counters.Threads.Current(),
					Sessions:    counters.Sessions.Current(),
					Users:       counters.Users.Current(),
					Warnings:    counters.Warnings.Current(),
					Requests:    counters.Requests.Current(),
					ActiveUsers: counters.ActiveUsers.Current(),
					Errors:      counters.Errors.Current(),
				})
			}
		}
	}()
	return nil
}

func NewProgress() *Progress {

	return &Progress{
		Values: make(map[string][]float64, 7),
	}
}

func (progess *Progress) pushValues(values Values) {
	progess.pushedCounter++

	termWidth, termHeight := termui.TerminalDimensions()

	maxData := termWidth/2 - 5
	if maxData < 5 {
		maxData = 5
	}

	if progess.pushedCounter > maxData {
		// need to wrap all arrays to size before adding new values
		for k, v := range progess.Values {
			lenV := len(v)
			if lenV > maxData {
				lenV = maxData
			}
			progess.Values[k] = v[1:lenV]
		}
	}

	progess.Values[Threads] = append(progess.Values[Threads], float64(values.Threads))
	progess.Values[Sessions] = append(progess.Values[Sessions], float64(values.Sessions))
	progess.Values[Users] = append(progess.Values[Users], float64(values.Users))
	progess.Values[Warnings] = append(progess.Values[Warnings], float64(values.Warnings))
	progess.Values[Requests] = append(progess.Values[Requests], float64(values.Requests))
	progess.Values[ActiveUsers] = append(progess.Values[ActiveUsers], float64(values.ActiveUsers))
	progess.Values[Errors] = append(progess.Values[Errors], float64(values.Errors))

	progess.Render(termWidth, termHeight)
}

var (
	RedYellowTheme = []termui.Color{
		termui.ColorRed,
		termui.ColorYellow,
	}

	DefaultTheme = []termui.Color{
		termui.ColorBlue,
		termui.ColorGreen,
		termui.ColorCyan,
		termui.ColorMagenta,
		termui.ColorYellow,
		termui.ColorWhite,
		termui.ColorRed,
	}
)

func (progress *Progress) Render(termWidth, termHeight int) {
	if progress.pushedCounter < 2 {
		return
	}

	grid := termui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		termui.NewRow(.5,
			termui.NewCol(.5, LineGraph("Throughput", []string{Requests}, [][]float64{progress.Values[Requests]}, termWidth/2, termHeight/2, DefaultTheme)),
			termui.NewCol(.5, LineGraph("Errors/Warnings", []string{Errors, Warnings}, [][]float64{progress.Values[Errors], progress.Values[Warnings]}, termWidth/2, termHeight/2, RedYellowTheme)),
		),
		termui.NewRow(.5,
			termui.NewCol(.5, LineGraph("Users", []string{Users, ActiveUsers}, [][]float64{progress.Values[Users], progress.Values[ActiveUsers]}, termWidth/2, termHeight/2, DefaultTheme)),
		),
	)

	termui.Render(grid)
}

func LineGraph(title string, labels []string, values [][]float64, maxX, maxY int, colors []termui.Color) *widgets.Plot {
	linegraph := widgets.NewPlot()
	linegraph.Title = title
	linegraph.Data = values
	linegraph.AxesColor = termui.ColorWhite
	linegraph.LineColors = colors
	linegraph.ShowAxes = false
	return linegraph
}

// func InvisibleTextBox(text string, color termui.Color, rect Rect) *widgets.List {
// 	// tb := widgets.NewParagraph()
// 	// tb.Text = text
// 	tb := widgets.NewList()
// 	tb.Border = false
// 	tb.TextStyle.Fg = color
// 	tb.SetRect(rect.x1, rect.y1, rect.x2, rect.y2)
// 	tb.Rows
// 	// tb.PaddingBottom = 0
// 	// tb.PaddingLeft = 0
// 	// tb.PaddingRight = 0
// 	// tb.PaddingTop = 0
// 	// tb.Min = image.Pt(rect.x1, rect.y1)
// 	return tb
// }
