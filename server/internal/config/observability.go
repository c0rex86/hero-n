package config

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "heroin_requests_total", Help: "Total requests"},
		[]string{"method", "endpoint", "status"},
	)
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "heroin_request_duration_seconds", Help: "Request duration"},
		[]string{"method", "endpoint"},
	)
	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "heroin_active_connections", Help: "Active connections"},
	)
)

func SetupObservability(ctx context.Context, cfg ObservabilityConfig) error {
	setupLogging(cfg)
	return setupMetrics(ctx, cfg)
}

func setupLogging(cfg ObservabilityConfig) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func setupMetrics(ctx context.Context, cfg ObservabilityConfig) error {
	if cfg.PrometheusAddr == "" { return nil }
	
	prometheus.MustRegister(RequestsTotal, RequestDuration, ActiveConnections)
	
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	server := &http.Server{Addr: cfg.PrometheusAddr, Handler: mux}
	
	go func() {
		slog.Info("prometheus metrics server starting", "addr", cfg.PrometheusAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("prometheus server failed", "error", err)
		}
	}()
	
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()
	
	return nil
}

func RecordRequest(method, endpoint, status string, duration float64) {
	RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}
