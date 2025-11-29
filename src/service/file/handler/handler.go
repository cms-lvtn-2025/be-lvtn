package handler

import (
	"context"
	"database/sql"
	"time"

	pb "thaily/proto/file"
	"thaily/src/service/pkg/logger"
)

type Handler struct {
	pb.UnimplementedFileServiceServer
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := h.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	queryWithArgs := logger.BuildQueryString(query, args...)
	logger.AddQueryToTrace(ctx, queryWithArgs, duration.Milliseconds())

	return row
}

func (h *Handler) query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := h.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	queryWithArgs := logger.BuildQueryString(query, args...)
	logger.AddQueryToTrace(ctx, queryWithArgs, duration.Milliseconds())

	return rows, err
}

func (h *Handler) execQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := h.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	queryWithArgs := logger.BuildQueryString(query, args...)
	logger.AddQueryToTrace(ctx, queryWithArgs, duration.Milliseconds())

	return result, err
}
