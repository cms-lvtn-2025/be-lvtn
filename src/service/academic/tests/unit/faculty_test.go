package unit

import (
	"context"
	"database/sql"
	"testing"
	pb "thaily/proto/academic"
	"thaily/src/service/academic/tests/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFaculty_Create(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	facultyMock := mocks.NewFacultyMock(db)

	t.Run("Success - Create faculty", func(t *testing.T) {
		// Mock UUID generation and INSERT query
		mock.ExpectExec("INSERT INTO Faculty").
			WithArgs(sqlmock.AnyArg(), "Computer Science", "admin").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Execute
		req := &pb.CreateFacultyRequest{
			Title:     "Computer Science",
			CreatedBy: "admin",
		}
		resp, err := facultyMock.CreateFaculty(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Faculty)
		assert.Equal(t, "Computer Science", resp.Faculty.Title)
		assert.Equal(t, "admin", resp.Faculty.CreatedBy)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Missing title", func(t *testing.T) {
		// Execute with empty title
		req := &pb.CreateFacultyRequest{
			Title:     "",
			CreatedBy: "admin",
		}
		resp, err := facultyMock.CreateFaculty(context.Background(), req)

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
		mock.ExpectExec("INSERT INTO Faculty").
			WithArgs(sqlmock.AnyArg(), "Computer Science", "admin").
			WillReturnError(sql.ErrConnDone)

		// Execute
		req := &pb.CreateFacultyRequest{
			Title:     "Computer Science",
			CreatedBy: "admin",
		}
		resp, err := facultyMock.CreateFaculty(context.Background(), req)

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

func TestFaculty_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	facultyMock := mocks.NewFacultyMock(db)

	t.Run("Success - List faculties", func(t *testing.T) {
		// Expected query result
		rows := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Computer Science", "admin").
			AddRow("2", "Mathematics", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Faculty").WillReturnRows(rows)

		// Execute
		req := &pb.ListFacultiesRequest{}
		resp, err := facultyMock.ListFaculties(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Faculties, 2)

		// Check first faculty
		firstFaculty := resp.Faculties[0]
		assert.Equal(t, "1", firstFaculty.Id)
		assert.Equal(t, "Computer Science", firstFaculty.Title)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database query fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Faculty").WillReturnError(sql.ErrConnDone)

		// Execute
		req := &pb.ListFacultiesRequest{}
		resp, err := facultyMock.ListFaculties(context.Background(), req)

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

func TestFaculty_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	facultyMock := mocks.NewFacultyMock(db)

	t.Run("Success - Faculty found", func(t *testing.T) {
		// Expected query result
		row := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Computer Science", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Faculty WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		// Execute
		req := &pb.GetFacultyRequest{Id: "1"}
		resp, err := facultyMock.GetFaculty(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Faculty)
		assert.Equal(t, "1", resp.Faculty.Id)
		assert.Equal(t, "Computer Science", resp.Faculty.Title)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Faculty not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Faculty WHERE id = ?").
			WithArgs("999").
			WillReturnError(sql.ErrNoRows)

		// Execute
		req := &pb.GetFacultyRequest{Id: "999"}
		resp, err := facultyMock.GetFaculty(context.Background(), req)

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
		req := &pb.GetFacultyRequest{Id: ""}
		resp, err := facultyMock.GetFaculty(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestFaculty_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	facultyMock := mocks.NewFacultyMock(db)

	t.Run("Success - Update faculty", func(t *testing.T) {
		mock.ExpectExec("UPDATE Faculty SET").
			WithArgs("Updated Computer Science", "admin", "1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Mock GetFaculty call after update
		row := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Updated Computer Science", "admin")
		mock.ExpectQuery("SELECT (.+) FROM Faculty WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		// Execute
		req := &pb.UpdateFacultyRequest{
			Id:        "1",
			Title:     &[]string{"Updated Computer Science"}[0],
			UpdatedBy: "admin",
		}
		resp, err := facultyMock.UpdateFaculty(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Faculty)
		assert.Equal(t, "Updated Computer Science", resp.Faculty.Title)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Faculty not found for update", func(t *testing.T) {
		mock.ExpectExec("UPDATE Faculty SET").
			WithArgs("Updated Computer Science", "admin", "999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Execute
		req := &pb.UpdateFacultyRequest{
			Id:        "999",
			Title:     &[]string{"Updated Computer Science"}[0],
			UpdatedBy: "admin",
		}
		resp, err := facultyMock.UpdateFaculty(context.Background(), req)

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

func TestFaculty_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	facultyMock := mocks.NewFacultyMock(db)

	t.Run("Success - Delete faculty", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Faculty WHERE id = ?").
			WithArgs("1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Execute
		req := &pb.DeleteFacultyRequest{Id: "1"}
		resp, err := facultyMock.DeleteFaculty(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)

		// Verify mock expectations
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Faculty not found for deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Faculty WHERE id = ?").
			WithArgs("999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Execute
		req := &pb.DeleteFacultyRequest{Id: "999"}
		resp, err := facultyMock.DeleteFaculty(context.Background(), req)

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
		req := &pb.DeleteFacultyRequest{Id: ""}
		resp, err := facultyMock.DeleteFaculty(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC status
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}
