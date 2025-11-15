package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// GraphQL Request Metrics
	GraphQLRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "graphql_requests_total",
			Help: "Total number of GraphQL requests",
		},
		[]string{"operation", "status"},
	)

	GraphQLRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "graphql_request_duration_seconds",
			Help:    "Duration of GraphQL requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// GraphQL Operation Metrics
	GraphQLOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "graphql_operations_total",
			Help: "Total number of GraphQL operations by type",
		},
		[]string{"operation_type", "operation_name", "status"},
	)

	GraphQLFieldResolutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "graphql_field_resolution_duration_seconds",
			Help:    "Duration of GraphQL field resolution",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"field_name", "type_name"},
	)

	// Business Logic Metrics
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "graphql_active_connections",
			Help: "Number of active GraphQL connections",
		},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Custom Business Metrics
	ThesisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "thesis_operations_total",
			Help: "Total number of thesis operations",
		},
		[]string{"operation", "status"},
	)

	UserActivitiesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_activities_total",
			Help: "Total number of user activities",
		},
		[]string{"activity_type", "user_role"},
	)

	FileOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_operations_total",
			Help: "Total number of file operations",
		},
		[]string{"operation", "file_type", "status"},
	)

	ServiceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health",
			Help: "Health status of services (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service"},
	)
)

// Record GraphQL request metrics
func RecordGraphQLRequest(operation, status string, duration time.Duration) {
	GraphQLRequestsTotal.WithLabelValues(operation, status).Inc()
	GraphQLRequestDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// Record GraphQL operation metrics
func RecordGraphQLOperation(operationType, operationName, status string) {
	GraphQLOperationsTotal.WithLabelValues(operationType, operationName, status).Inc()
}

// Record field resolution metrics
func RecordFieldResolution(fieldName, typeName string, duration time.Duration) {
	GraphQLFieldResolutionDuration.WithLabelValues(fieldName, typeName).Observe(duration.Seconds())
}

// Record HTTP request metrics
func RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// Record business metrics
func RecordThesisOperation(operation, status string) {
	ThesisOperationsTotal.WithLabelValues(operation, status).Inc()
}

func RecordUserActivity(activityType, userRole string) {
	UserActivitiesTotal.WithLabelValues(activityType, userRole).Inc()
}

func RecordFileOperation(operation, fileType, status string) {
	FileOperationsTotal.WithLabelValues(operation, fileType, status).Inc()
}

// Set service health
func SetServiceHealth(service string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	ServiceHealth.WithLabelValues(service).Set(value)
}

// Update active connections
func SetActiveConnections(count int) {
	ActiveConnections.Set(float64(count))
}

// HTTP Middleware for request metrics
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Capture response status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		statusCode := strconv.Itoa(wrapped.statusCode)

		RecordHTTPRequest(r.Method, r.URL.Path, statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// Health check endpoint
func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"graphql-gateway"}`))
	}
}
