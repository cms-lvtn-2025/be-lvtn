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

func TestMajor_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	majorMock := mocks.NewMajorMock(db)

	t.Run("Success - Create major", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO Major").
			WithArgs(sqlmock.AnyArg(), "Computer Science", "CSC", "admin").
			WillReturnResult(sqlmock.NewResult(1, 1))

		req := &pb.CreateMajorRequest{
			Title:       "Computer Science",
			FacultyCode: "CSC",
			CreatedBy:   "admin",
		}
		resp, err := majorMock.CreateMajor(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Major)
		assert.Equal(t, "Computer Science", resp.Major.Title)
		assert.Equal(t, "CSC", resp.Major.FacultyCode)
		assert.Equal(t, "admin", resp.Major.CreatedBy)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Missing title", func(t *testing.T) {
		req := &pb.CreateMajorRequest{
			Title:       "",
			FacultyCode: "CSC",
			CreatedBy:   "admin",
		}
		resp, err := majorMock.CreateMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "title is required")
	})

	t.Run("Error - Missing faculty code", func(t *testing.T) {
		req := &pb.CreateMajorRequest{
			Title:       "Computer Science",
			FacultyCode: "",
			CreatedBy:   "admin",
		}
		resp, err := majorMock.CreateMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "faculty_code is required")
	})

	t.Run("Error - Database insert fails", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO Major").
			WithArgs(sqlmock.AnyArg(), "Computer Science", "CSC", "admin").
			WillReturnError(sql.ErrConnDone)

		req := &pb.CreateMajorRequest{
			Title:       "Computer Science",
			FacultyCode: "CSC",
			CreatedBy:   "admin",
		}
		resp, err := majorMock.CreateMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMajor_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	majorMock := mocks.NewMajorMock(db)

	t.Run("Success - List majors", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "faculty_code", "created_by"}).
			AddRow("1", "Computer Science", "CSC", "admin").
			AddRow("2", "Mathematics", "MATH", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Major").WillReturnRows(rows)

		req := &pb.ListMajorsRequest{}
		resp, err := majorMock.ListMajors(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Majors, 2)

		firstMajor := resp.Majors[0]
		assert.Equal(t, "1", firstMajor.Id)
		assert.Equal(t, "Computer Science", firstMajor.Title)
		assert.Equal(t, "CSC", firstMajor.FacultyCode)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database query fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Major").WillReturnError(sql.ErrConnDone)

		req := &pb.ListMajorsRequest{}
		resp, err := majorMock.ListMajors(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMajor_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	majorMock := mocks.NewMajorMock(db)

	t.Run("Success - Major found", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "title", "faculty_code", "created_by"}).
			AddRow("1", "Computer Science", "CSC", "admin")

		mock.ExpectQuery("SELECT (.+) FROM Major WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		req := &pb.GetMajorRequest{Id: "1"}
		resp, err := majorMock.GetMajor(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Major)
		assert.Equal(t, "1", resp.Major.Id)
		assert.Equal(t, "Computer Science", resp.Major.Title)
		assert.Equal(t, "CSC", resp.Major.FacultyCode)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Major not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM Major WHERE id = ?").
			WithArgs("999").
			WillReturnError(sql.ErrNoRows)

		req := &pb.GetMajorRequest{Id: "999"}
		resp, err := majorMock.GetMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ID", func(t *testing.T) {
		req := &pb.GetMajorRequest{Id: ""}
		resp, err := majorMock.GetMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestMajor_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	majorMock := mocks.NewMajorMock(db)

	t.Run("Success - Update major", func(t *testing.T) {
		mock.ExpectExec("UPDATE Major SET").
			WithArgs("Computer Science Updated", "CSC2", "admin", "1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		row := sqlmock.NewRows([]string{"id", "title", "faculty_code", "created_by"}).
			AddRow("1", "Computer Science Updated", "CSC2", "admin")
		mock.ExpectQuery("SELECT (.+) FROM Major WHERE id = ?").
			WithArgs("1").
			WillReturnRows(row)

		req := &pb.UpdateMajorRequest{
			Id:          "1",
			Title:       &[]string{"Computer Science Updated"}[0],
			FacultyCode: &[]string{"CSC2"}[0],
			UpdatedBy:   "admin",
		}
		resp, err := majorMock.UpdateMajor(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Major)
		assert.Equal(t, "Computer Science Updated", resp.Major.Title)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Major not found for update", func(t *testing.T) {
		mock.ExpectExec("UPDATE Major SET").
			WithArgs("Computer Science Updated", "CSC2", "admin", "999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		req := &pb.UpdateMajorRequest{
			Id:          "999",
			Title:       &[]string{"Computer Science Updated"}[0],
			FacultyCode: &[]string{"CSC2"}[0],
			UpdatedBy:   "admin",
		}
		resp, err := majorMock.UpdateMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMajor_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	majorMock := mocks.NewMajorMock(db)

	t.Run("Success - Delete major", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Major WHERE id = ?").
			WithArgs("1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		req := &pb.DeleteMajorRequest{Id: "1"}
		resp, err := majorMock.DeleteMajor(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Major not found for deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM Major WHERE id = ?").
			WithArgs("999").
			WillReturnResult(sqlmock.NewResult(0, 0))

		req := &pb.DeleteMajorRequest{Id: "999"}
		resp, err := majorMock.DeleteMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ID", func(t *testing.T) {
		req := &pb.DeleteMajorRequest{Id: ""}
		resp, err := majorMock.DeleteMajor(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}
