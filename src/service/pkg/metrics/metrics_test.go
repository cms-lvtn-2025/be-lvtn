package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRecordGrpcRequest(t *testing.T) {
	t.Run("Record successful gRPC request", func(t *testing.T) {
		service := "test-service"
		method := "TestMethod"
		status := "success"
		duration := time.Millisecond * 100

		RecordGrpcRequest(service, method, status, duration)

		value := testutil.ToFloat64(GrpcRequestsTotal.WithLabelValues(service, method, status))
		assert.Greater(t, value, float64(0))
	})
}

func TestRecordDatabaseQuery(t *testing.T) {
	t.Run("Record database query metrics", func(t *testing.T) {
		service := "academic"
		operation := "select"
		status := "success"
		duration := time.Millisecond * 50

		RecordDatabaseQuery(service, operation, status, duration)

		value := testutil.ToFloat64(DatabaseQueriesTotal.WithLabelValues(service, operation, status))
		assert.Greater(t, value, float64(0))
	})
}

func TestSetServiceHealth(t *testing.T) {
	t.Run("Set service healthy", func(t *testing.T) {
		serviceName := "test-service"
		SetServiceHealth(serviceName, true)
		value := testutil.ToFloat64(ServiceHealth.WithLabelValues(serviceName))
		assert.Equal(t, float64(1), value)
	})
}

func TestActiveConnections(t *testing.T) {
	t.Run("Increment active connections", func(t *testing.T) {
		serviceName := "test-service-connections"
		initialValue := testutil.ToFloat64(GrpcActiveConnections.WithLabelValues(serviceName))
		IncActiveConnections(serviceName)
		newValue := testutil.ToFloat64(GrpcActiveConnections.WithLabelValues(serviceName))
		assert.Equal(t, initialValue+1, newValue)
	})
}
