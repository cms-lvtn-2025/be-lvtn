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

func TestSemester_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	semesterMock := mocks.NewSemesterMock(db)

	t.Run("Success - Create semester", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO Semester").
			WithArgs(sqlmock.AnyArg(), "Fall 2024", "admin").
			WillReturnResult(sqlmock.NewResult(1, 1))

		req := &pb.CreateSemesterRequest{
			Title:     "Fall 2024",
			CreatedBy: "admin",
		}
		resp, err := semesterMock.CreateSemester(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Semester)
		assert.Equal(t, "Fall 2024", resp.Semester.Title)
		assert.Equal(t, "admin", resp.Semester.CreatedBy)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Missing title", func(t *testing.T) {
		req := &pb.CreateSemesterRequest{
			Title:     "",
			CreatedBy: "admin",
		}
		resp, err := semesterMock.CreateSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "title is required")
	})

	t.Run("Error - Database insert fails", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO Semester").
			WithArgs(sqlmock.AnyArg(), "Fall 2024", "admin").
			WillReturnError(sql.ErrConnDone)

		req := &pb.CreateSemesterRequest{
			Title:     "Fall 2024",
			CreatedBy: "admin",
		}
		resp, err := semesterMock.CreateSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSemester_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	semesterMock := mocks.NewSemesterMock(db)

	t.Run("Success - List semesters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Fall 2024", "admin").
			AddRow("2", "Spring 2024", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Semester").WillReturnRows(rows)

		req := &pb.ListSemestersRequest{}
		resp, err := semesterMock.ListSemesters(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Semesters, 2)

		firstSemester := resp.Semesters[0]
		assert.Equal(t, "1", firstSemester.Id)
		assert.Equal(t, "Fall 2024", firstSemester.Title)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database query fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Semester").WillReturnError(sql.ErrConnDone)

		req := &pb.ListSemestersRequest{}
		resp, err := semesterMock.ListSemesters(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSemester_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	semesterMock := mocks.NewSemesterMock(db)

	t.Run("Success - Semester found", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Fall 2024", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Semester WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		req := &pb.GetSemesterRequest{Id: "1"}
		resp, err := semesterMock.GetSemester(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Semester)
		assert.Equal(t, "1", resp.Semester.Id)
		assert.Equal(t, "Fall 2024", resp.Semester.Title)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Semester not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Semester WHERE id = ?").
			WithArgs("999").
			WillReturnError(sql.ErrNoRows)

		req := &pb.GetSemesterRequest{Id: "999"}
		resp, err := semesterMock.GetSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ID", func(t *testing.T) {
		req := &pb.GetSemesterRequest{Id: ""}
		resp, err := semesterMock.GetSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestSemester_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	semesterMock := mocks.NewSemesterMock(db)

	t.Run("Success - Update semester", func(t *testing.T) {
		mock.ExpectExec("UPDATE Semester SET").
			WithArgs("Fall 2024 Updated", "admin", "1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		row := sqlmock.NewRows([]string{"id", "title", "created_by"}).
			AddRow("1", "Fall 2024 Updated", "admin")
		mock.ExpectQuery("SELECT (.+) FROM Semester WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		req := &pb.UpdateSemesterRequest{
			Id:        "1",
			Title:     &[]string{"Fall 2024 Updated"}[0],
			UpdatedBy: "admin",
		}
		resp, err := semesterMock.UpdateSemester(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Semester)
		assert.Equal(t, "Fall 2024 Updated", resp.Semester.Title)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Semester not found for update", func(t *testing.T) {
		mock.ExpectExec("UPDATE Semester SET").
			WithArgs("Fall 2024 Updated", "admin", "999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		req := &pb.UpdateSemesterRequest{
			Id:        "999",
			Title:     &[]string{"Fall 2024 Updated"}[0],
			UpdatedBy: "admin",
		}
		resp, err := semesterMock.UpdateSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSemester_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	semesterMock := mocks.NewSemesterMock(db)

	t.Run("Success - Delete semester", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Semester WHERE id = ?").
			WithArgs("1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		req := &pb.DeleteSemesterRequest{Id: "1"}
		resp, err := semesterMock.DeleteSemester(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Semester not found for deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Semester WHERE id = ?").
			WithArgs("999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		req := &pb.DeleteSemesterRequest{Id: "999"}
		resp, err := semesterMock.DeleteSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ID", func(t *testing.T) {
		req := &pb.DeleteSemesterRequest{Id: ""}
		resp, err := semesterMock.DeleteSemester(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}
