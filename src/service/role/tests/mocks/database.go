package mocks

import (
	"testing"

	"thaily/src/service/pkg/metrics"

	"github.com/DATA-DOG/go-sqlmock"
)

// NewMockDB creates a new mock database connection
func NewMockDB(t *testing.T) (*metrics.DatabaseHandler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	dbHandler := metrics.NewDatabaseHandler(db, "role")
	return dbHandler, mock
}
