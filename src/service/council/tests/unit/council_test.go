package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"thaily/src/service/council/handler"
	"thaily/src/service/council/tests/fixtures"

	pb "thaily/proto/council"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ==================== CREATE COUNCIL TESTS ====================

func TestCreateCouncil_Success(t *testing.T) {
	// Create mock DB
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateCouncilRequest()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Council").
		WithArgs(
			sqlmock.AnyArg(), // id
			req.Title,
			req.MajorCode,
			req.SemesterCode,
			req.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT query (GetCouncil is called after insert)
	mock.ExpectQuery("SELECT (.+) FROM Council WHERE id = ?").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			"test-id", req.Title, req.MajorCode, req.SemesterCode,
			time.Now(), time.Now(), req.CreatedBy, "",
		))

	// Execute
	resp, err := h.CreateCouncil(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Council)
	assert.Equal(t, req.Title, resp.Council.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCouncil_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := &pb.CreateCouncilRequest{
		Title:        "",
		MajorCode:    "CS",
		SemesterCode: "2024-1",
		CreatedBy:    "test-user",
	}

	// Execute
	resp, err := h.CreateCouncil(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateCouncil_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateCouncilRequest()

	// Mock INSERT with error
	mock.ExpectExec("INSERT INTO Council").
		WillReturnError(sql.ErrConnDone)

	// Execute
	resp, err := h.CreateCouncil(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GET COUNCIL TESTS ====================

func TestGetCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	council := fixtures.GetTestCouncil()

	// Mock SELECT query - order must match handler: id, title, major_code, semester_code, created_at, updated_at, created_by, updated_by
	mock.ExpectQuery("SELECT (.+) FROM Council WHERE id = ?").
		WithArgs(council.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			council.Id, council.Title, council.MajorCode, council.SemesterCode,
			council.CreatedAt.AsTime(), council.UpdatedAt.AsTime(), council.CreatedBy, council.UpdatedBy,
		))

	// Execute
	resp, err := h.GetCouncil(context.Background(), &pb.GetCouncilRequest{Id: council.Id})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Council)
	assert.Equal(t, council.Id, resp.Council.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouncil_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock SELECT with no rows
	mock.ExpectQuery("SELECT (.+) FROM Council WHERE id = ?").
		WithArgs(fixtures.TestCouncilID1).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.GetCouncil(context.Background(), &pb.GetCouncilRequest{Id: fixtures.TestCouncilID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouncil_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.GetCouncil(context.Background(), &pb.GetCouncilRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== UPDATE COUNCIL TESTS ====================

func TestUpdateCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateCouncilRequest()

	// Mock UPDATE - handler builds dynamic query with all fields
	mock.ExpectExec("UPDATE Council SET").
		WithArgs(
			*req.Title,
			*req.MajorCode,
			*req.SemesterCode,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT for GetCouncil (called after update)
	mock.ExpectQuery("SELECT (.+) FROM Council WHERE id = ?").
		WithArgs(req.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			req.Id, *req.Title, *req.MajorCode, *req.SemesterCode,
			time.Now(), time.Now(), "test-user", req.UpdatedBy,
		))

	// Execute
	resp, err := h.UpdateCouncil(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Council)
	assert.Equal(t, *req.Title, resp.Council.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCouncil_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateCouncilRequest()

	// Mock UPDATE
	mock.ExpectExec("UPDATE Council SET").
		WithArgs(
			*req.Title,
			*req.MajorCode,
			*req.SemesterCode,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT that returns NotFound
	mock.ExpectQuery("SELECT (.+) FROM Council WHERE id = ?").
		WithArgs(req.Id).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.UpdateCouncil(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code()) // Handler returns Internal error when GetCouncil fails
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCouncil_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	title := "Updated"
	req := &pb.UpdateCouncilRequest{
		Id:        "",
		Title:     &title,
		UpdatedBy: "test-user",
	}

	// Execute
	resp, err := h.UpdateCouncil(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== DELETE COUNCIL TESTS ====================

func TestDeleteCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE
	mock.ExpectExec("DELETE FROM Council WHERE id = ?").
		WithArgs(fixtures.TestCouncilID1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	resp, err := h.DeleteCouncil(context.Background(), &pb.DeleteCouncilRequest{Id: fixtures.TestCouncilID1})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCouncil_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE with no rows
	mock.ExpectExec("DELETE FROM Council WHERE id = ?").
		WithArgs(fixtures.TestCouncilID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	resp, err := h.DeleteCouncil(context.Background(), &pb.DeleteCouncilRequest{Id: fixtures.TestCouncilID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCouncil_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.DeleteCouncil(context.Background(), &pb.DeleteCouncilRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== LIST COUNCILS TESTS ====================

func TestListCouncils_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	council1 := fixtures.GetTestCouncil()
	council2 := fixtures.GetTestCouncil2()

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Council").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock SELECT - order: id, title, major_code, semester_code, created_at, updated_at, created_by, updated_by
	mock.ExpectQuery("SELECT (.+) FROM Council").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).
			AddRow(
				council1.Id, council1.Title, council1.MajorCode, council1.SemesterCode,
				council1.CreatedAt.AsTime(), council1.UpdatedAt.AsTime(), council1.CreatedBy, council1.UpdatedBy,
			).
			AddRow(
				council2.Id, council2.Title, council2.MajorCode, council2.SemesterCode,
				council2.CreatedAt.AsTime(), council2.UpdatedAt.AsTime(), council2.CreatedBy, council2.UpdatedBy,
			))

	// Execute
	req := fixtures.GetTestListCouncilsRequest()
	resp, err := h.ListCouncils(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Councils, 2)
	assert.Equal(t, int32(2), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListCouncils_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Council").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock SELECT - order: id, title, major_code, semester_code, created_at, updated_at, created_by, updated_by
	mock.ExpectQuery("SELECT (.+) FROM Council").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}))

	// Execute
	req := fixtures.GetTestListCouncilsRequest()
	resp, err := h.ListCouncils(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Councils, 0)
	assert.Equal(t, int32(0), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListCouncils_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT with error
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Council").
		WillReturnError(sql.ErrConnDone)

	// Execute
	req := fixtures.GetTestListCouncilsRequest()
	resp, err := h.ListCouncils(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}
