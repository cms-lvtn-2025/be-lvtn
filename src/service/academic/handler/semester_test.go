package handler

import (
	"context"
	"database/sql"
	"testing"
	pb "thaily/proto/academic"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CreateSemester(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewTestHandler(db)

	t.Run("Success - Create semester", func(t *testing.T) {
		// Mock UUID generation and INSERT query
		mock.ExpectExec("INSERT INTO Semester").
			WithArgs(sqlmock.AnyArg(), "Fall 2024", "admin").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Execute
		req := &pb.CreateSemesterRequest{
			Title:     "Fall 2024",
			CreatedBy: "admin",
		}
		resp, err := handler.CreateSemester(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Semester)
		assert.NotEmpty(t, resp.Semester.Id)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Missing title", func(t *testing.T) {
		// Execute with missing title
		req := &pb.CreateSemesterRequest{
			Title:     "",
			CreatedBy: "admin",
		}
		resp, err := handler.CreateSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "title is required")
	})

	t.Run("Error - Database insert fails", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO Semester").
			WithArgs(sqlmock.AnyArg(), "Fall 2024", "admin").
			WillReturnError(sql.ErrConnDone)

		// Execute
		req := &pb.CreateSemesterRequest{
			Title:     "Fall 2024",
			CreatedBy: "admin",
		}
		resp, err := handler.CreateSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestHandler_ListSemesters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewTestHandler(db)

	t.Run("Success - List semesters", func(t *testing.T) {
		// Expected query result
		rows := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Fall 2024", "admin").
			AddRow("2", "Spring 2024", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Semester").WillReturnRows(rows)

		// Execute
		req := &pb.ListSemestersRequest{}
		resp, err := handler.ListSemesters(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Semesters, 2)

		// Check first semester
		firstSemester := resp.Semesters[0]
		assert.Equal(t, "1", firstSemester.Id)
		assert.Equal(t, "Fall 2024", firstSemester.Title)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database query fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Semester").WillReturnError(sql.ErrConnDone)

		// Execute
		req := &pb.ListSemestersRequest{}
		resp, err := handler.ListSemesters(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestHandler_GetSemester(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewTestHandler(db)

	t.Run("Success - Semester found", func(t *testing.T) {
		// Expected query result
		row := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Fall 2024", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Semester WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		// Execute
		req := &pb.GetSemesterRequest{Id: "1"}
		resp, err := handler.GetSemester(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Semester)
		assert.Equal(t, "1", resp.Semester.Id)
		assert.Equal(t, "Fall 2024", resp.Semester.Title)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Semester not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Semester WHERE id = ?").
			WithArgs("999").
			WillReturnError(sql.ErrNoRows)

		// Execute
		req := &pb.GetSemesterRequest{Id: "999"}
		resp, err := handler.GetSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ID", func(t *testing.T) {
		// Execute with empty ID
		req := &pb.GetSemesterRequest{Id: ""}
		resp, err := handler.GetSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestHandler_UpdateSemester(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewTestHandler(db)

	t.Run("Success - Update semester", func(t *testing.T) {
		mock.ExpectExec("UPDATE Semester SET").
			WithArgs("Fall 2024 Updated", "admin", "1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Mock GetSemester call after update
		row := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Fall 2024 Updated", "admin")
		mock.ExpectQuery("SELECT (.+) FROM Semester WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		// Execute
		req := &pb.UpdateSemesterRequest{
			Id:        "1",
			Title:     &[]string{"Fall 2024 Updated"}[0],
			UpdatedBy: "admin",
		}
		resp, err := handler.UpdateSemester(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Semester)
		assert.Equal(t, "Fall 2024 Updated", resp.Semester.Title)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Semester not found for update", func(t *testing.T) {
		mock.ExpectExec("UPDATE Semester SET").
			WithArgs("Fall 2024 Updated", "admin", "999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Execute
		req := &pb.UpdateSemesterRequest{
			Id:        "999",
			Title:     &[]string{"Fall 2024 Updated"}[0],
			UpdatedBy: "admin",
		}
		resp, err := handler.UpdateSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestHandler_DeleteSemester(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewTestHandler(db)

	t.Run("Success - Delete semester", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Semester WHERE id = ?").
			WithArgs("1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Execute
		req := &pb.DeleteSemesterRequest{Id: "1"}
		resp, err := handler.DeleteSemester(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Semester not found for deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Semester WHERE id = ?").
			WithArgs("999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Execute
		req := &pb.DeleteSemesterRequest{Id: "999"}
		resp, err := handler.DeleteSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ID", func(t *testing.T) {
		// Execute with empty ID
		req := &pb.DeleteSemesterRequest{Id: ""}
		resp, err := handler.DeleteSemester(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}
