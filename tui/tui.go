package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
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
		Updaters      []func() error
		pushedCounter int
	}
)

const (
	// MaxX = 300

	Threads     = "threads"
	Sessions    = "sessions"
	Users       = "users"
	Warnings    = "warnings"
	Requests    = "requests"
	ActiveUsers = "activeusers"
	Errors      = "errors"

	DefaultAxesColor   = cell.ColorNavy
	DefaultLabelColor  = cell.ColorGreen
	DefaultBorderColor = cell.ColorWhite
)

func StartProgressTui(ctx context.Context, cancel func(), counters *statistics.ExecutionCounters) error {
	if counters == nil {
		return errors.Errorf("no statistics counters provided")
	}

	progress := NewProgress()

	throughputGraph, err := progress.LineGraph("Throughput", []string{Requests}, []cell.Color{cell.ColorCyan})
	if err != nil {
		return err
	}
	errWarnGraph, err := progress.LineGraph("Errors/Warnings", []string{Errors, Warnings}, []cell.Color{cell.ColorRed, cell.ColorYellow})
	if err != nil {
		return err
	}
	sessionsGraph, err := progress.LineGraph("Current sessions", []string{ActiveUsers}, []cell.Color{cell.ColorCyan})
	if err != nil {
		return err
	}

	countersList, err := progress.List("Counters",
		[]string{Threads, Sessions, Users, ActiveUsers, Requests, Warnings, Errors},
		[]cell.Color{cell.ColorCyan, cell.ColorCyan, cell.ColorCyan, cell.ColorCyan, cell.ColorCyan, cell.ColorYellow, cell.ColorRed},
	)
	if err != nil {
		return err
	}

	layout := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(throughputGraph),
					container.Right(errWarnGraph),
				),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(sessionsGraph),
					container.Right(countersList...),
				),
			),
		),
	}

	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		return err
	}

	c, err := container.New(t, container.ID("root"))
	if err != nil {
		return err
	}

	c.Update("root", layout...)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.Key('q') || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	// TODO handle Run error
	go termdash.Run(ctx, t, c, termdash.RedrawInterval(time.Second), termdash.KeyboardSubscriber(quitter))

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second): // TODO make a ticker
				progress.pushValues(Values{
					Threads:     counters.Threads.Current(),
					Sessions:    counters.Sessions.Current(),
					Users:       counters.Users.Current(),
					Warnings:    counters.Warnings.Current(),
					Requests:    counters.Requests.Current(),
					ActiveUsers: counters.ActiveUsers.Current(),
					Errors:      counters.Errors.Current(),
				})
				if err := progress.Update(); err != nil {
					// TODO handle error
				}
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

func (progess *Progress) Update() error {
	var mErr *multierror.Error
	for _, updater := range progess.Updaters {
		if err := updater(); err != nil {
			multierror.Append(err)
		}
	}
	return helpers.FlattenMultiError(mErr)
}

func (progess *Progress) pushValues(values Values) {
	progess.pushedCounter++

	// if progess.pushedCounter > MaxX {
	// 	// need to wrap all arrays to size before adding new values
	// 	for k, v := range progess.Values {
	// 		lenV := len(v)
	// 		if lenV > MaxX {
	// 			lenV = MaxX
	// 		}
	// 		progess.Values[k] = v[1:lenV]
	// 	}
	// }

	progess.Values[Threads] = append(progess.Values[Threads], float64(values.Threads))
	progess.Values[Sessions] = append(progess.Values[Sessions], float64(values.Sessions))
	progess.Values[Users] = append(progess.Values[Users], float64(values.Users))
	progess.Values[Warnings] = append(progess.Values[Warnings], float64(values.Warnings))
	progess.Values[Requests] = append(progess.Values[Requests], float64(values.Requests))
	progess.Values[ActiveUsers] = append(progess.Values[ActiveUsers], float64(values.ActiveUsers))
	progess.Values[Errors] = append(progess.Values[Errors], float64(values.Errors))
}

func (progress *Progress) LineGraph(title string, labels []string, colors []cell.Color) (container.Option, error) {

	if len(labels) != len(colors) {
		return nil, errors.Errorf("labels<%d> and colors<%d> most have same length", len(labels), len(colors))
	}

	linegraph, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(DefaultAxesColor)),
		linechart.YLabelCellOpts(cell.FgColor(DefaultLabelColor)),
		linechart.XLabelCellOpts(cell.FgColor(DefaultLabelColor)),
	)
	if err != nil {
		return nil, err
	}

	if progress.Updaters == nil {
		progress.Updaters = make([]func() error, 0, 1)
	}

	legendWidget, err := text.New()

	progress.Updaters = append(progress.Updaters, func() error {
		var mErr *multierror.Error

		legendWidget.Reset()
		for i, label := range labels {
			if err := linegraph.Series(label, progress.Values[label], linechart.SeriesCellOpts(cell.FgColor(colors[i]))); err != nil {
				multierror.Append(err)
			}
			legendText := ""
			if i > 0 {
				legendText = " "
			}
			legendText += fmt.Sprintf("%s %0.2f", label, progress.Values[label][len(progress.Values[label])-1])
			if err := legendWidget.Write(legendText, text.WriteCellOpts(cell.FgColor(colors[i]))); err != nil {
				multierror.Append(err)
			}
		}
		return helpers.FlattenMultiError(mErr)
	})

	containerOption := container.SplitHorizontal(
		container.Top(container.PlaceWidget(linegraph), container.Border(linestyle.Light), container.BorderTitle(title), container.BorderColor(DefaultBorderColor)),
		container.Bottom(container.PlaceWidget(legendWidget)),
		container.SplitPercent(98),
	)

	return containerOption, err
}

func (progress *Progress) List(title string, labels []string, colors []cell.Color) ([]container.Option, error) {
	if len(labels) != len(colors) {
		return nil, errors.Errorf("labels<%d> and colors<%d> most have same length", len(labels), len(colors))
	}

	listWidget, err := text.New()
	if err != nil {
		return nil, err
	}
	containerOptions := []container.Option{
		container.PlaceWidget(listWidget),
		container.Border(linestyle.Light),
		container.BorderTitle(title),
	}

	if progress.Updaters == nil {
		progress.Updaters = make([]func() error, 0, 1)
	}
	progress.Updaters = append(progress.Updaters, func() error {
		listWidget.Reset()
		for i, label := range labels {
			listText := ""
			if i > 0 {
				listText += "\n"
			}
			listText += fmt.Sprintf("%s %0.2f", label, progress.Values[label][len(progress.Values[label])-1])
			if err := listWidget.Write(listText, text.WriteCellOpts(cell.FgColor(colors[i]))); err != nil {
				multierror.Append(err)
			}
		}
		return nil
	})
	return containerOptions, nil
}
