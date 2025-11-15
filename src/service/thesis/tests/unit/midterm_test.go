package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"thaily/src/service/thesis/handler"
	"thaily/src/service/thesis/tests/fixtures"

	common "thaily/proto/common"
	pb "thaily/proto/thesis"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestCreateMidterm_Success tests successful midterm creation
func TestCreateMidterm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Midterm").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.MidtermTitle1,
			fixtures.MidtermGrade1,
			"submitted", // status enum as string
			fixtures.MidtermFeedback1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetMidterm
	rows := sqlmock.NewRows([]string{
		"id", "title", "grade", "status", "feedback",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.MidtermID1,
		fixtures.MidtermTitle1,
		fixtures.MidtermGrade1,
		"submitted",
		fixtures.MidtermFeedback1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Midterm WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateMidtermRequest{
		Title:     fixtures.MidtermTitle1,
		Grade:     &fixtures.MidtermGrade1,
		Status:    pb.MidtermStatus_SUBMITTED,
		Feedback:  &fixtures.MidtermFeedback1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateMidterm(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Midterm)
	assert.Equal(t, fixtures.MidtermTitle1, resp.Midterm.Title)
}

// TestCreateMidterm_MissingTitle tests creation with missing title
func TestCreateMidterm_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateMidtermRequest{
		Title:     "",
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateMidterm_DuplicateEntry tests duplicate midterm creation
func TestCreateMidterm_DuplicateEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO Midterm").
		WithArgs(
			sqlmock.AnyArg(),
			fixtures.MidtermTitle1,
			fixtures.MidtermGrade1,
			"submitted",
			fixtures.MidtermFeedback1,
			fixtures.CreatedBy,
		).
		WillReturnError(sql.ErrNoRows)

	req := &pb.CreateMidtermRequest{
		Title:     fixtures.MidtermTitle1,
		Grade:     &fixtures.MidtermGrade1,
		Status:    pb.MidtermStatus_SUBMITTED,
		Feedback:  &fixtures.MidtermFeedback1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestGetMidterm_Success tests successful midterm retrieval
func TestGetMidterm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "title", "grade", "status", "feedback",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.MidtermID1,
		fixtures.MidtermTitle1,
		fixtures.MidtermGrade1,
		"submitted",
		fixtures.MidtermFeedback1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Midterm WHERE id").
		WithArgs(fixtures.MidtermID1).
		WillReturnRows(rows)

	req := &pb.GetMidtermRequest{
		Id: fixtures.MidtermID1,
	}

	resp, err := h.GetMidterm(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Midterm)
	assert.Equal(t, fixtures.MidtermID1, resp.Midterm.Id)
	assert.Equal(t, fixtures.MidtermTitle1, resp.Midterm.Title)
	assert.Equal(t, pb.MidtermStatus_SUBMITTED, resp.Midterm.Status)
}

// TestGetMidterm_NotFound tests midterm not found
func TestGetMidterm_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Midterm WHERE id").
		WithArgs(fixtures.MidtermID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetMidtermRequest{
		Id: fixtures.MidtermID1,
	}

	resp, err := h.GetMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetMidterm_MissingID tests getting midterm with missing ID
func TestGetMidterm_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetMidtermRequest{
		Id: "",
	}

	resp, err := h.GetMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateMidterm_Success tests successful midterm update
// Note: Handler uses dynamic UPDATE - only includes non-nil fields
func TestUpdateMidterm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Midterm Title"
	newGrade := int32(95)
	newStatus := pb.MidtermStatus_PASS

	// Mock UPDATE query - only title, grade, status are provided (feedback is nil)
	// Args: title, grade, status, updated_by, id
	mock.ExpectExec("UPDATE Midterm SET").
		WithArgs(
			newTitle,
			newGrade,
			"pass", // status enum as string
			fixtures.UpdatedBy,
			fixtures.MidtermID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetMidterm
	rows := sqlmock.NewRows([]string{
		"id", "title", "grade", "status", "feedback",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.MidtermID1,
		newTitle,
		newGrade,
		"pass",
		fixtures.MidtermFeedback1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Midterm WHERE id").
		WithArgs(fixtures.MidtermID1).
		WillReturnRows(rows)

	req := &pb.UpdateMidtermRequest{
		Id:     fixtures.MidtermID1,
		Title:  &newTitle,
		Grade:  &newGrade,
		Status: &newStatus,
		// Feedback is nil - handler won't include it in UPDATE
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateMidterm(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Midterm)
	assert.Equal(t, newTitle, resp.Midterm.Title)
	assert.Equal(t, newGrade, resp.Midterm.Grade)
	assert.Equal(t, pb.MidtermStatus_PASS, resp.Midterm.Status)
}

// TestUpdateMidterm_NotFound tests updating non-existent midterm
func TestUpdateMidterm_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Title"

	mock.ExpectExec("UPDATE Midterm SET").
		WithArgs(
			newTitle,
			fixtures.UpdatedBy,
			fixtures.MidtermID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Midterm WHERE id").
		WithArgs(fixtures.MidtermID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateMidtermRequest{
		Id:        fixtures.MidtermID1,
		Title:     &newTitle,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateMidterm_NoFieldsToUpdate tests update with no fields
func TestUpdateMidterm_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateMidtermRequest{
		Id:        fixtures.MidtermID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteMidterm_Success tests successful midterm deletion
func TestDeleteMidterm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Midterm WHERE id").
		WithArgs(fixtures.MidtermID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteMidtermRequest{
		Id: fixtures.MidtermID1,
	}

	resp, err := h.DeleteMidterm(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteMidterm_NotFound tests deleting non-existent midterm
func TestDeleteMidterm_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Midterm WHERE id").
		WithArgs(fixtures.MidtermID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteMidtermRequest{
		Id: fixtures.MidtermID1,
	}

	resp, err := h.DeleteMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteMidterm_MissingID tests deletion with missing ID
func TestDeleteMidterm_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteMidtermRequest{
		Id: "",
	}

	resp, err := h.DeleteMidterm(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListMidterms_Success tests successful midterms listing
func TestListMidterms_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock COUNT query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	// Mock SELECT query
	rows := sqlmock.NewRows([]string{
		"id", "title", "grade", "status", "feedback",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.MidtermID1,
			fixtures.MidtermTitle1,
			fixtures.MidtermGrade1,
			"submitted",
			fixtures.MidtermFeedback1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.MidtermID2,
			fixtures.MidtermTitle2,
			fixtures.MidtermGrade2,
			"pass",
			fixtures.MidtermFeedback2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Midterm").
		WillReturnRows(rows)

	req := &pb.ListMidtermsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListMidterms(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.Midterms, 2)
	assert.Equal(t, fixtures.MidtermTitle1, resp.Midterms[0].Title)
	assert.Equal(t, pb.MidtermStatus_SUBMITTED, resp.Midterms[0].Status)
}

// TestListMidterms_Empty tests listing with no results
func TestListMidterms_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "title", "grade", "status", "feedback",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Midterm").
		WillReturnRows(rows)

	req := &pb.ListMidtermsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListMidterms(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.Midterms)
}
