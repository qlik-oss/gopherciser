package enigmainterceptors

import (
	"context"

	"github.com/qlik-oss/enigma-go/v3"
)

type (
	MetricsHandler struct {
		Log func(invocation *enigma.Invocation, metrics *enigma.InvocationMetrics, response *enigma.InvocationResponse)
	}
)

func (m *MetricsHandler) MetricsInterceptor(ctx context.Context, invocation *enigma.Invocation, proceed enigma.InterceptorContinuation) *enigma.InvocationResponse {
	ctxWithMetrics, metricsCollector := enigma.WithMetricsCollector(ctx)
	response := proceed(ctxWithMetrics, invocation)
	metrics := metricsCollector.Metrics()
	m.Log(invocation, metrics, response)

	return response
}
