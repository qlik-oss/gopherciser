// +build js

package buildmetrics

import (
	"context"
	"time"
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
func ReportApiResult(path string, responseCode int, duration time.Duration) {
	return
}

// ReportSuccess shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func ReportSuccess(action string, label string, time float64) {
	return
}

// ReportFailure shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func ReportFailure(action string, label string) {
	return
}

// ReportError shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func ReportError(action string, label string) {
	return
}

// ReportWarning shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func ReportWarning(action string, label string) {
	return
}

// AddUser shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func AddUser() {
	return
}

// RemoveUser shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func RemoveUser() {
	return
}

// PullMetrics shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func PullMetrics(ctx context.Context, metricsPort int, registeredActions []string) error {
	return nil
}

// PushMetrics shall never report Prometheus metrics for WASM/JS builds as it is not supported nor wanted, hence "return"
// Implemented as a way to dynamically import prometheus
func PushMetrics(ctx context.Context, metricsPort int, metricsAddress string, metricsLabel string, registeredActions []string) error {
	return nil
}
