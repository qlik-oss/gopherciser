package metrics

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

func setupMetrics(actions []string) error {
	prometheus.MustRegister(GopherActions)
	prometheus.MustRegister(GopherWarnings)
	prometheus.MustRegister(GopherErrors)
	prometheus.MustRegister(GopherUsersTotal)
	prometheus.MustRegister(GopherActiveUsers)
	prometheus.MustRegister(GopherResponseTimes)
	prometheus.MustRegister(GopherActionLatencyHist)

	err := gopherRegistry.Register(GopherActions)
	if err != nil {
		return err
	}
	err = gopherRegistry.Register(GopherWarnings)
	if err != nil {
		return err
	}
	err = gopherRegistry.Register(GopherErrors)
	if err != nil {
		return err
	}
	err = gopherRegistry.Register(GopherUsersTotal)
	if err != nil {
		return err
	}
	err = gopherRegistry.Register(GopherActiveUsers)
	if err != nil {
		return err
	}
	err = gopherRegistry.Register(GopherResponseTimes)
	if err != nil {
		return err
	}
	err = gopherRegistry.Register(GopherActionLatencyHist)
	if err != nil {
		return err
	}

	// Initialize metrics
	for _, action := range actions {
		GopherActions.WithLabelValues("success", action).Add(0)
		GopherActions.WithLabelValues("failure", action).Add(0)
	}

	return nil
}

// PushMetrics handles the constant pushing of metrics to prometheus
func PushMetrics(ctx context.Context, metricsPort int, metricsAddress, job string, groupingKeys, actions []string) error {
	err := setupMetrics(actions)
	if err != nil {
		return err
	}

	var addr = flag.String("push-address", fmt.Sprintf("%s:%d", metricsAddress, metricsPort), "The address to push prometheus metrics")
	pusher := push.New(*addr, job).Gatherer(gopherRegistry)
	for _, gk := range groupingKeys {
		kv := strings.SplitN(gk, "=", 2)
		if len(kv) < 2 || len(kv[0]) == 0 {
			return fmt.Errorf("can't parse grouping key %q: must be in 'key=value' form", gk)
		}
		pusher = pusher.Grouping(kv[0], kv[1])
	}

	//Pushes prometheus metrics every minute
	const interval time.Duration = 1 * time.Minute
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := pusher.Push()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Push error received: %s", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

// PullMetrics handle the serving of prometheus metrics on the metrics endpoint
func PullMetrics(ctx context.Context, metricsPort int, actions []string) error {
	err := setupMetrics(actions)
	if err != nil {
		return err
	}

	var addr = flag.String("pull-address", fmt.Sprintf(":%d", metricsPort), "The address to listen on for HTTP requests.")
	srv := &http.Server{Addr: *addr}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			_, _ = fmt.Fprintf(os.Stderr, "Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	go func() {
		<-ctx.Done()
		//nolint:staticcheck
		if err := srv.Shutdown(context.Background()); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Httpserver: Shutdown() error: %s", err)
		}
	}()

	return nil
}
