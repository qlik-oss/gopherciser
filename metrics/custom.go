package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	promNS = "gopherciser"
)

// GopherActions action counter
var GopherActions = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: promNS,
		Name:      "actions_total",
		Help:      "Number of gopherciser actions and their result. DOES NOT USE LABEL!",
	},
	[]string{"result", "action"},
)

// GopherWarnings execution Warnings
var GopherWarnings = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: promNS,
		Name:      "warnings_total",
		Help:      "Number of gopherciser execution warnings per action/label.",
	},
	[]string{"action"},
)

// GopherErrors execution Errors
var GopherErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: promNS,
		Name:      "errors_per_action",
		Help:      "Number of gopherciser execution errors per action/label and app.",
	},
	[]string{"action"},
)

// GopherUsersTotal simulated users
var GopherUsersTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: promNS,
	Name:      "users_total",
	Help:      "Number of gopherciser users simulated.",
})

// GopherActiveUsers currently active
var GopherActiveUsers = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: promNS,
	Name:      "active_users",
	Help:      "Current amount of active users",
})

// GopherResponseTimes response times
var GopherResponseTimes = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace:  promNS,
		Name:       "response_times",
		Help:       "Summarized response times per action/label",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
	[]string{"action"},
)

// GopherActionLatencyHist is a histogram tracking the response times of actions
var GopherActionLatencyHist = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: promNS,
		Name:      "response_times_seconds",
		Help:      "latency of actions/label",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 4, 6},
	},
	[]string{"action"},
)

// BuildInfo -
var BuildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: promNS,
		Name:      "build_info",
		Help:      "A constant metric labeled with build information for " + promNS,
	},
	[]string{"version", "revision", "goversion", "goarch", "goos"},
)

//GopherRegistry registers the metrics in a registry to be used for prometheus push
var gopherRegistry = prometheus.NewRegistry()
