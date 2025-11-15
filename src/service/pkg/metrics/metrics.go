package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// gRPC metrics
var (
	// Request total counter
	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	// Request duration histogram
	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// Active connections gauge
	GrpcActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_active_connections",
			Help: "Number of active gRPC connections",
		},
		[]string{"service"},
	)

	// Database metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"service", "operation", "status"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"service", "operation"},
	)

	// Service health
	ServiceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health",
			Help: "Health status of service (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service"},
	)
)

// Helper functions for recording metrics
func RecordGrpcRequest(service, method, status string, duration time.Duration) {
	GrpcRequestsTotal.WithLabelValues(service, method, status).Inc()
	GrpcRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

func RecordDatabaseQuery(service, operation, status string, duration time.Duration) {
	DatabaseQueriesTotal.WithLabelValues(service, operation, status).Inc()
	DatabaseQueryDuration.WithLabelValues(service, operation).Observe(duration.Seconds())
}

func SetServiceHealth(service string, healthy bool) {
	if healthy {
		ServiceHealth.WithLabelValues(service).Set(1)
	} else {
		ServiceHealth.WithLabelValues(service).Set(0)
	}
}

func IncActiveConnections(service string) {
	GrpcActiveConnections.WithLabelValues(service).Inc()
}

func DecActiveConnections(service string) {
	GrpcActiveConnections.WithLabelValues(service).Dec()
}
