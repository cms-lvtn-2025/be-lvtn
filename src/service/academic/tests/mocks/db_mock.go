package mocks

import (
	"context"
	"database/sql"
)

// Test helper functions and mocks

// MockHandler for testing academic service operations
type MockHandler struct {
	db *sql.DB
}

// NewMockHandler creates a new mock handler instance for testing
func NewMockHandler(db *sql.DB) *MockHandler {
	return &MockHandler{db: db}
}

// Helper methods for testing database operations
func (h *MockHandler) ExecQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return h.db.ExecContext(ctx, query, args...)
}

func (h *MockHandler) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return h.db.QueryRowContext(ctx, query, args...)
}

func (h *MockHandler) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return h.db.QueryContext(ctx, query, args...)
}

// Database connection interface for testing
type DBInterface interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}
