package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/qlik-oss/gopherciser/statistics"
)

const (
	graphMaxSamples = 7200 // two hours at one sample per second
	graphLabelWidth = 8
)

// blocks are eighth-step vertical bars, index 0 is empty.
var blocks = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type (
	// series describes one graphed counter.
	series struct {
		name  string
		color color.Color
		rate  bool // graph change per second instead of the counter value
		read  func(*statistics.ExecutionCounters) uint64
	}

	// sample is a snapshot of all series counters.
	sample struct {
		time   time.Time
		values []uint64
	}

	// graphView plots sampled execution counters as stacked bar charts with
	// scales adjusted to the visible values.
	graphView struct {
		series  []series
		samples []sample
	}
)

var execSeries = []series{
	{"Active users", colorTitle, false, func(c *statistics.ExecutionCounters) uint64 { return c.ActiveUsers.Current() }},
	{"Sessions", colorOK, false, func(c *statistics.ExecutionCounters) uint64 { return c.Sessions.Current() }},
	{"Requests/s", colorKey, true, func(c *statistics.ExecutionCounters) uint64 { return c.Requests.Current() }},
	{"Errors", colorError, false, func(c *statistics.ExecutionCounters) uint64 { return c.Errors.Current() }},
	{"Warnings", colorWarning, false, func(c *statistics.ExecutionCounters) uint64 { return c.Warnings.Current() }},
}

func newGraphView() *graphView { return &graphView{series: execSeries} }

// reset drops all samples.
func (g *graphView) reset() { g.samples = nil }

// record takes a sample of the counters.
func (g *graphView) record(counters *statistics.ExecutionCounters, now time.Time) {
	values := make([]uint64, len(g.series))
	for i, s := range g.series {
		values[i] = s.read(counters)
	}
	g.samples = append(g.samples, sample{time: now, values: values})
	if extra := len(g.samples) - graphMaxSamples; extra > 0 {
		g.samples = g.samples[extra:]
	}
}

// points returns the values to plot for series i, at most n and newest last.
func (g *graphView) points(i, n int) []float64 {
	s := g.series[i]
	from := 0
	if s.rate {
		from = 1
	}
	if extra := len(g.samples) - from - n; extra > 0 {
		from += extra
	}
	points := make([]float64, 0, n)
	for j := from; j < len(g.samples); j++ {
		value := float64(g.samples[j].values[i])
		if s.rate {
			prev := g.samples[j-1]
			seconds := g.samples[j].time.Sub(prev.time).Seconds()
			if seconds <= 0 {
				seconds = 1
			}
			value = (value - float64(prev.values[i])) / seconds
		}
		points = append(points, value)
	}
	return points
}

// view draws the charts in w by h cells.
func (g *graphView) view(w, h int) string {
	if w < graphLabelWidth+5 || h < 2 {
		return ""
	}
	if len(g.samples) < 1 {
		return center(styleDisabled.Render("no samples yet"), w, h)
	}

	// each chart gets a title row plus an equal share of the plot rows
	rows := h / len(g.series)
	count := len(g.series)
	if rows < 2 {
		rows = 2
		count = h / rows
	}
	plotWidth := w - graphLabelWidth
	span := g.samples[len(g.samples)-1].time.Sub(g.samples[0].time).Truncate(time.Second)

	var lines []string
	for i := 0; i < count; i++ {
		lines = append(lines, g.chart(i, plotWidth, rows-1, span)...)
	}
	return strings.Join(lines, "\n")
}

// chart draws series i as a title line followed by the plot rows.
func (g *graphView) chart(i, plotWidth, plotHeight int, span time.Duration) []string {
	s := g.series[i]
	points := g.points(i, plotWidth)

	var current, peak float64
	if len(points) > 0 {
		current = points[len(points)-1]
	}
	for _, p := range points {
		peak = math.Max(peak, p)
	}
	top := niceCeil(peak)

	key := styleKey.Render
	lines := []string{fmt.Sprintf("%s %s %s %s %s %s %s", fg(s.color).Bold(true).Render(s.name),
		key("now"), formatValue(current, s.rate), key("max"), formatValue(peak, s.rate), key("window"), span)}

	// y axis labels on the left, bars fill from the left and scroll once the plot is full
	style := fg(s.color)
	levels := plotHeight * (len(blocks) - 1)
	for row := 0; row < plotHeight; row++ {
		label := ""
		switch row {
		case 0:
			label = formatValue(top, s.rate)
		case plotHeight - 1:
			label = "0"
		}
		label = fmt.Sprintf("%*s ", graphLabelWidth-1, label)

		bars := make([]rune, 0, len(points))
		for _, p := range points {
			filled := 0
			if top > 0 {
				filled = int(math.Round(p / top * float64(levels)))
			}
			if p > 0 && filled < 1 {
				filled = 1
			}
			// rows are drawn top down, the bottom row holds the first eighths
			rowLevels := filled - (plotHeight-1-row)*(len(blocks)-1)
			bars = append(bars, blocks[min(max(rowLevels, 0), len(blocks)-1)])
		}
		lines = append(lines, styleDisabled.Render(label)+style.Render(string(bars)))
	}
	return lines
}

// niceCeil rounds value up to 1, 2 or 5 times a power of ten for a readable axis.
func niceCeil(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value <= 1 {
		return 1
	}
	magnitude := math.Pow(10, math.Floor(math.Log10(value)))
	for _, step := range []float64{1, 2, 5, 10} {
		if nice := step * magnitude; nice >= value {
			return nice
		}
	}
	return 10 * magnitude
}

// formatValue prints a value compactly, with a decimal for small rates.
func formatValue(value float64, rate bool) string {
	switch {
	case rate && value < 10 && value != math.Trunc(value):
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", value), "0"), ".")
	case value >= 1e6:
		return fmt.Sprintf("%.1fM", value/1e6)
	case value >= 1e4:
		return fmt.Sprintf("%.1fk", value/1e3)
	default:
		return fmt.Sprintf("%.0f", value)
	}
}
