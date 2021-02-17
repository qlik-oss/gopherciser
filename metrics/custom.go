package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	promNS = "gopherciser"
)

// ApiCallDuration histogram for API call duration
var ApiCallDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "api_request_duration_seconds",
		Help:    "A histogram of HTTP request durations.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.15, 0.25, 0.5, 1., 2.0, 2.5, 5., 10.},
	},
	[]string{"path", "method", "status_code"},
)

// ApiCallDurationQuantile summary for API call duration
var ApiCallDurationQuantile = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "api_request_duration_quantiles_seconds",
		Help:       "A summary of HTTP request durations",
		Objectives: map[float64]float64{0.1: 0.1, 0.5: 0.05, 0.95: 0.01, 0.99: 0.001, 0.999: 0.0001},
	},
	[]string{"path", "method", "status_code"},
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
