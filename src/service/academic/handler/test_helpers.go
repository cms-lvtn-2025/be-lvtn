package handler

import (
	"context"
	"database/sql"
)

// Test helper functions and mocks

// Mock Handler for testing
type TestHandler struct {
	db *sql.DB
}

// NewTestHandler creates a new handler instance for testing
func NewTestHandler(db *sql.DB) *TestHandler {
	return &TestHandler{db: db}
}

// Helper methods for testing - these would normally be in the actual handler
func (h *TestHandler) execQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return h.db.ExecContext(ctx, query, args...)
}

func (h *TestHandler) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return h.db.QueryRowContext(ctx, query, args...)
}

func (h *TestHandler) query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return h.db.QueryContext(ctx, query, args...)
}
