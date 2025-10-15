package handler

import (
	"context"
	"database/sql"

	pb "thaily/proto/thesis"
)

type Handler struct {
	pb.UnimplementedThesisServiceServer
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return h.db.QueryRowContext(ctx, query, args...)
}

func (h *Handler) query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return h.db.QueryContext(ctx, query, args...)
}

func (h *Handler) execQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return h.db.ExecContext(ctx, query, args...)
}
