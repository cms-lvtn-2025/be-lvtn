package metrics

import (
	"context"
	"database/sql"
	"time"
)

// DatabaseHandler wraps database operations with metrics
type DatabaseHandler struct {
	db          *sql.DB
	serviceName string
}

// NewDatabaseHandler creates a new database handler with metrics
func NewDatabaseHandler(db *sql.DB, serviceName string) *DatabaseHandler {
	return &DatabaseHandler{
		db:          db,
		serviceName: serviceName,
	}
}

// QueryRowContext executes a query with metrics
func (h *DatabaseHandler) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := h.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics
	RecordDatabaseQuery(h.serviceName, "select", "success", duration)

	return row
}

// QueryContext executes a query with metrics
func (h *DatabaseHandler) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := h.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Determine status
	status := "success"
	if err != nil {
		status = "error"
	}

	// Record metrics
	RecordDatabaseQuery(h.serviceName, "select", status, duration)

	return rows, err
}

// ExecContext executes a statement with metrics
func (h *DatabaseHandler) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := h.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Determine operation type
	operation := "update"
	if len(query) > 6 {
		switch query[:6] {
		case "INSERT", "insert":
			operation = "insert"
		case "UPDATE", "update":
			operation = "update"
		case "DELETE", "delete":
			operation = "delete"
		}
	}

	// Determine status
	status := "success"
	if err != nil {
		status = "error"
	}

	// Record metrics
	RecordDatabaseQuery(h.serviceName, operation, status, duration)

	return result, err
}

// GetDB returns the underlying database connection
func (h *DatabaseHandler) GetDB() *sql.DB {
	return h.db
}
