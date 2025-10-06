package handler

import (
	"context"
	"database/sql"
	"time"
	pb "thaily/proto/user"
	"thaily/src/pkg/logger"
)

type Handler struct {
	pb.UnimplementedUserServiceServer
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// Helper methods
func (h *Handler) getDB() *sql.DB {
	return h.db
}

func (h *Handler) execQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := h.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Add query to trace
	logger.AddQueryToTrace(ctx, query, duration.Milliseconds())

	return result, err
}

func (h *Handler) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := h.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Add query to trace
	logger.AddQueryToTrace(ctx, query, duration.Milliseconds())

	return row
}

func (h *Handler) query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := h.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Add query to trace
	logger.AddQueryToTrace(ctx, query, duration.Milliseconds())

	return rows, err
}
