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

// TestCreateGradeReview_Success tests successful grade review creation
func TestCreateGradeReview_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Grade_review").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.GradeReviewTitle1,
			fixtures.GradeReviewGrade1,
			fixtures.GradeReviewTeacherCode1,
			"pending", // status enum as string
			fixtures.GradeReviewNotes1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetGradeReview
	rows := sqlmock.NewRows([]string{
		"id", "title", "review_grade", "teacher_code", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.GradeReviewID1,
		fixtures.GradeReviewTitle1,
		fixtures.GradeReviewGrade1,
		fixtures.GradeReviewTeacherCode1,
		"pending",
		fixtures.GradeReviewNotes1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Grade_review WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateGradeReviewRequest{
		Title:       fixtures.GradeReviewTitle1,
		ReviewGrade: &fixtures.GradeReviewGrade1,
		TeacherCode: fixtures.GradeReviewTeacherCode1,
		Status:      pb.FinalStatus_PENDING,
		Notes:       &fixtures.GradeReviewNotes1,
		CreatedBy:   fixtures.CreatedBy,
	}

	resp, err := h.CreateGradeReview(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.GradeReview)
	assert.Equal(t, fixtures.GradeReviewTitle1, resp.GradeReview.Title)
}

// TestCreateGradeReview_MissingTitle tests creation with missing title
func TestCreateGradeReview_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateGradeReviewRequest{
		Title:       "",
		TeacherCode: fixtures.GradeReviewTeacherCode1,
		CreatedBy:   fixtures.CreatedBy,
	}

	resp, err := h.CreateGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateGradeReview_MissingTeacherCode tests creation with missing teacher code
func TestCreateGradeReview_MissingTeacherCode(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateGradeReviewRequest{
		Title:       fixtures.GradeReviewTitle1,
		TeacherCode: "",
		CreatedBy:   fixtures.CreatedBy,
	}

	resp, err := h.CreateGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestGetGradeReview_Success tests successful grade review retrieval
func TestGetGradeReview_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "title", "review_grade", "teacher_code", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.GradeReviewID1,
		fixtures.GradeReviewTitle1,
		fixtures.GradeReviewGrade1,
		fixtures.GradeReviewTeacherCode1,
		"pending",
		fixtures.GradeReviewNotes1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Grade_review WHERE id").
		WithArgs(fixtures.GradeReviewID1).
		WillReturnRows(rows)

	req := &pb.GetGradeReviewRequest{
		Id: fixtures.GradeReviewID1,
	}

	resp, err := h.GetGradeReview(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.GradeReview)
	assert.Equal(t, fixtures.GradeReviewID1, resp.GradeReview.Id)
	assert.Equal(t, fixtures.GradeReviewTitle1, resp.GradeReview.Title)
	assert.Equal(t, pb.FinalStatus_PENDING, resp.GradeReview.Status)
}

// TestGetGradeReview_NotFound tests grade review not found
func TestGetGradeReview_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Grade_review WHERE id").
		WithArgs(fixtures.GradeReviewID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetGradeReviewRequest{
		Id: fixtures.GradeReviewID1,
	}

	resp, err := h.GetGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetGradeReview_MissingID tests getting grade review with missing ID
func TestGetGradeReview_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetGradeReviewRequest{
		Id: "",
	}

	resp, err := h.GetGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateGradeReview_Success tests successful grade review update
// Note: Handler uses dynamic UPDATE - only includes non-nil fields
func TestUpdateGradeReview_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Grade Review Title"
	newGrade := int32(95)
	newStatus := pb.FinalStatus_PASSED

	// Mock UPDATE query - only title, review_grade, status are provided
	// Args: title, review_grade, status, updated_by, id
	mock.ExpectExec("UPDATE GradeReview SET").
		WithArgs(
			newTitle,
			newGrade,
			"passed", // status enum as string
			fixtures.UpdatedBy,
			fixtures.GradeReviewID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetGradeReview
	rows := sqlmock.NewRows([]string{
		"id", "title", "review_grade", "teacher_code", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.GradeReviewID1,
		newTitle,
		newGrade,
		fixtures.GradeReviewTeacherCode1,
		"passed",
		fixtures.GradeReviewNotes1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Grade_review WHERE id").
		WithArgs(fixtures.GradeReviewID1).
		WillReturnRows(rows)

	req := &pb.UpdateGradeReviewRequest{
		Id:          fixtures.GradeReviewID1,
		Title:       &newTitle,
		ReviewGrade: &newGrade,
		Status:      &newStatus,
		// TeacherCode and Notes are nil - handler won't include them in UPDATE
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateGradeReview(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.GradeReview)
	assert.Equal(t, newTitle, resp.GradeReview.Title)
	assert.NotNil(t, resp.GradeReview.ReviewGrade)
	assert.Equal(t, newGrade, *resp.GradeReview.ReviewGrade)
	assert.Equal(t, pb.FinalStatus_PASSED, resp.GradeReview.Status)
}

// TestUpdateGradeReview_NotFound tests updating non-existent grade review
func TestUpdateGradeReview_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Title"

	mock.ExpectExec("UPDATE GradeReview SET").
		WithArgs(
			newTitle,
			fixtures.UpdatedBy,
			fixtures.GradeReviewID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Grade_review WHERE id").
		WithArgs(fixtures.GradeReviewID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateGradeReviewRequest{
		Id:        fixtures.GradeReviewID1,
		Title:     &newTitle,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateGradeReview_NoFieldsToUpdate tests update with no fields
func TestUpdateGradeReview_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateGradeReviewRequest{
		Id:        fixtures.GradeReviewID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteGradeReview_Success tests successful grade review deletion
func TestDeleteGradeReview_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Grade_review WHERE id").
		WithArgs(fixtures.GradeReviewID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteGradeReviewRequest{
		Id: fixtures.GradeReviewID1,
	}

	resp, err := h.DeleteGradeReview(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteGradeReview_NotFound tests deleting non-existent grade review
func TestDeleteGradeReview_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Grade_review WHERE id").
		WithArgs(fixtures.GradeReviewID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteGradeReviewRequest{
		Id: fixtures.GradeReviewID1,
	}

	resp, err := h.DeleteGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteGradeReview_MissingID tests deletion with missing ID
func TestDeleteGradeReview_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteGradeReviewRequest{
		Id: "",
	}

	resp, err := h.DeleteGradeReview(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListGradeReviews_Success tests successful grade reviews listing
func TestListGradeReviews_Success(t *testing.T) {
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
		"id", "title", "review_grade", "teacher_code", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.GradeReviewID1,
			fixtures.GradeReviewTitle1,
			fixtures.GradeReviewGrade1,
			fixtures.GradeReviewTeacherCode1,
			"pending",
			fixtures.GradeReviewNotes1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.GradeReviewID2,
			fixtures.GradeReviewTitle2,
			fixtures.GradeReviewGrade2,
			fixtures.GradeReviewTeacherCode2,
			"passed",
			fixtures.GradeReviewNotes2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Grade_review").
		WillReturnRows(rows)

	req := &pb.ListGradeReviewsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListGradeReviews(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.GradeReviews, 2)
	assert.Equal(t, fixtures.GradeReviewTitle1, resp.GradeReviews[0].Title)
	assert.Equal(t, pb.FinalStatus_PENDING, resp.GradeReviews[0].Status)
}

// TestListGradeReviews_Empty tests listing with no results
func TestListGradeReviews_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "title", "review_grade", "teacher_code", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Grade_review").
		WillReturnRows(rows)

	req := &pb.ListGradeReviewsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListGradeReviews(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.GradeReviews)
}
