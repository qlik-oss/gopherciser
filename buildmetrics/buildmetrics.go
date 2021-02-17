// +build linux darwin windows

package buildmetrics

import (
	"context"
	"strconv"
	"time"

	"github.com/qlik-oss/gopherciser/metrics"
)

var (
	enabled bool
)

// MetricEnabled returns whether metrics are enabled
func metricEnabled() bool {
	return enabled
}

func getLabel(action string, label string) string {
	if label != "" {
		return label
	}
	return action
}

// ReportApiResult reports the duration for a specific API path and response code
func ReportApiResult(action, label, path, method string, responseCode int, duration time.Duration) {
	if metricEnabled() {
		actionlabel := getLabel(action, label)
		resultString := strconv.Itoa(responseCode)
		metrics.ApiCallDuration.WithLabelValues(actionlabel, path, method, resultString).Observe(duration.Seconds())
		metrics.ApiCallDurationQuantile.WithLabelValues(actionlabel, path, method, resultString).Observe(duration.Seconds())
	}
}

// ReportSuccess is invoked when a simulated user action is successfully completed.
// This then updates Prometheus metrics correlating to this (ReponseTimes | Latency | success counter for an action)
func ReportSuccess(action string, label string, time float64) {
	if metricEnabled() {
		actionlabel := getLabel(action, label)
		metrics.GopherResponseTimes.WithLabelValues(actionlabel).Observe(time)
		metrics.GopherActionLatencyHist.WithLabelValues(actionlabel).Observe(time)
		metrics.GopherActions.WithLabelValues("success", actionlabel).Inc()
	}
}

// ReportFailure is invoked when a simulated user action fails.
// This then updates Prometheus metrics correlating to this (Failure counter for an action)
func ReportFailure(action string, label string) {
	if metricEnabled() {
		actionlabel := getLabel(action, label)
		metrics.GopherActions.WithLabelValues("failure", actionlabel).Inc()
	}
}

// ReportError is invoked when an error occurs in execution. A user action can in theory have many errors.
// This then updates Prometheus metrics correlating to this (Error counter)
func ReportError(action string, label string) {
	if metricEnabled() {
		actionlabel := getLabel(action, label)
		metrics.GopherErrors.WithLabelValues(actionlabel).Inc()
	}
}

// ReportWarning is invoked when an warning occurs in execution. A user action can in theory have many warnings.
// This then updates Prometheus metrics correlating to this (Warning counter)
func ReportWarning(action string, label string) {
	if metricEnabled() {
		actionlabel := getLabel(action, label)
		metrics.GopherWarnings.WithLabelValues(actionlabel).Inc()
	}
}

// AddUser is invoked when a new simulated user is added.
// This then updates Prometheus metrics correlating to this (Users total | Active users)
func AddUser() {
	if metricEnabled() {
		metrics.GopherUsersTotal.Inc()
		metrics.GopherActiveUsers.Inc()
	}
}

// RemoveUser is invoked when a new simulated user is added.
// This then updates Prometheus metrics correlating to this (Decrements Active users).
// Note Total users is for obvious reasons not decremented
func RemoveUser() {
	if metricEnabled() {
		metrics.GopherActiveUsers.Dec()
	}
}

// PullMetrics is called once to setup and enable Prometheus pull metrics on a certain endpoint
func PullMetrics(ctx context.Context, metricsPort int, registeredActions []string) error {
	enabled = true
	err := metrics.PullMetrics(ctx, metricsPort, registeredActions)
	if err != nil {
		return err
	}
	return nil
}

// PushMetrics is called once to setup and enable Prometheus push metrics to specified address
func PushMetrics(ctx context.Context, metricsPort int, metricsAddress, job string, groupingKeys, registeredActions []string) error {
	enabled = true
	err := metrics.PushMetrics(ctx, metricsPort, metricsAddress, job, groupingKeys, registeredActions)
	if err != nil {
		return err
	}
	return nil
}
