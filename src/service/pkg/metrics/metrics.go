package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// Request metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method"},
	)

	grpcRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_requests_in_flight",
			Help: "Number of gRPC requests currently being processed",
		},
		[]string{"service"},
	)

	// Database metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"service", "operation"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"service", "operation"},
	)

	// Error metrics
	grpcErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_errors_total",
			Help: "Total number of gRPC errors",
		},
		[]string{"service", "method", "code"},
	)

	// Service info
	serviceInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_info",
			Help: "Service information",
		},
		[]string{"service", "version"},
	)
)

// MetricsServer runs the Prometheus metrics HTTP server
type MetricsServer struct {
	serviceName string
	port        string
	server      *http.Server
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(serviceName, port string) *MetricsServer {
	return &MetricsServer{
		serviceName: serviceName,
		port:        port,
	}
}

// Start starts the metrics HTTP server
func (m *MetricsServer) Start() error {
	// Set service info
	serviceInfo.WithLabelValues(m.serviceName, "1.0.0").Set(1)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	m.server = &http.Server{
		Addr:    ":" + m.port,
		Handler: mux,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return nil
}

// Stop stops the metrics server
func (m *MetricsServer) Stop() error {
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.server.Shutdown(ctx)
	}
	return nil
}

// UnaryServerInterceptor returns a gRPC interceptor for metrics
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Track in-flight requests
		grpcRequestsInFlight.WithLabelValues(serviceName).Inc()
		defer grpcRequestsInFlight.WithLabelValues(serviceName).Dec()

		// Record start time
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Record duration
		duration := time.Since(start).Seconds()
		grpcRequestDuration.WithLabelValues(serviceName, info.FullMethod).Observe(duration)

		// Record request count
		statusCode := "OK"
		if err != nil {
			st, _ := status.FromError(err)
			statusCode = st.Code().String()
			grpcErrorsTotal.WithLabelValues(serviceName, info.FullMethod, statusCode).Inc()
		}
		grpcRequestsTotal.WithLabelValues(serviceName, info.FullMethod, statusCode).Inc()

		return resp, err
	}
}

// RecordDBQuery records a database query metric
func RecordDBQuery(serviceName, operation string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(serviceName, operation).Inc()
	dbQueryDuration.WithLabelValues(serviceName, operation).Observe(duration.Seconds())
}
