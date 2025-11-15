package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns HTTP handler for Prometheus metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartMetricsServer starts HTTP server for metrics on specified port
func StartMetricsServer(port string) error {
	http.Handle("/metrics", Handler())
	return http.ListenAndServe(":"+port, nil)
}
