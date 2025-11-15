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

// TestCreateFinal_Success tests successful final creation
func TestCreateFinal_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Final").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.FinalTitle1,
			fixtures.FinalSupervisorGrade1,
			fixtures.FinalDepartmentGrade1,
			fixtures.FinalGrade1,
			"passed", // status enum as string
			fixtures.FinalNotes1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetFinal
	rows := sqlmock.NewRows([]string{
		"id", "title", "supervisor_grade", "department_grade", "final_grade", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FinalID1,
		fixtures.FinalTitle1,
		fixtures.FinalSupervisorGrade1,
		fixtures.FinalDepartmentGrade1,
		fixtures.FinalGrade1,
		"passed",
		fixtures.FinalNotes1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Final WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateFinalRequest{
		Title:           fixtures.FinalTitle1,
		SupervisorGrade: &fixtures.FinalSupervisorGrade1,
		DepartmentGrade: &fixtures.FinalDepartmentGrade1,
		FinalGrade:      &fixtures.FinalGrade1,
		Status:          pb.FinalStatus_PASSED,
		Notes:           &fixtures.FinalNotes1,
		CreatedBy:       fixtures.CreatedBy,
	}

	resp, err := h.CreateFinal(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Final)
	assert.Equal(t, fixtures.FinalTitle1, resp.Final.Title)
}

// TestCreateFinal_MissingTitle tests creation with missing title
func TestCreateFinal_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateFinalRequest{
		Title:     "",
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateFinal_DuplicateEntry tests duplicate final creation
func TestCreateFinal_DuplicateEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO Final").
		WithArgs(
			sqlmock.AnyArg(),
			fixtures.FinalTitle1,
			fixtures.FinalSupervisorGrade1,
			fixtures.FinalDepartmentGrade1,
			fixtures.FinalGrade1,
			"passed",
			fixtures.FinalNotes1,
			fixtures.CreatedBy,
		).
		WillReturnError(sql.ErrNoRows)

	req := &pb.CreateFinalRequest{
		Title:           fixtures.FinalTitle1,
		SupervisorGrade: &fixtures.FinalSupervisorGrade1,
		DepartmentGrade: &fixtures.FinalDepartmentGrade1,
		FinalGrade:      &fixtures.FinalGrade1,
		Status:          pb.FinalStatus_PASSED,
		Notes:           &fixtures.FinalNotes1,
		CreatedBy:       fixtures.CreatedBy,
	}

	resp, err := h.CreateFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestGetFinal_Success tests successful final retrieval
func TestGetFinal_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "title", "supervisor_grade", "department_grade", "final_grade", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FinalID1,
		fixtures.FinalTitle1,
		fixtures.FinalSupervisorGrade1,
		fixtures.FinalDepartmentGrade1,
		fixtures.FinalGrade1,
		"passed",
		fixtures.FinalNotes1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Final WHERE id").
		WithArgs(fixtures.FinalID1).
		WillReturnRows(rows)

	req := &pb.GetFinalRequest{
		Id: fixtures.FinalID1,
	}

	resp, err := h.GetFinal(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Final)
	assert.Equal(t, fixtures.FinalID1, resp.Final.Id)
	assert.Equal(t, fixtures.FinalTitle1, resp.Final.Title)
	assert.Equal(t, pb.FinalStatus_PASSED, resp.Final.Status)
}

// TestGetFinal_NotFound tests final not found
func TestGetFinal_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Final WHERE id").
		WithArgs(fixtures.FinalID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetFinalRequest{
		Id: fixtures.FinalID1,
	}

	resp, err := h.GetFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetFinal_MissingID tests getting final with missing ID
func TestGetFinal_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetFinalRequest{
		Id: "",
	}

	resp, err := h.GetFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateFinal_Success tests successful final update
// Note: Handler uses dynamic UPDATE - only includes non-nil fields
func TestUpdateFinal_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Final Title"
	newSupervisorGrade := int32(95)
	newStatus := pb.FinalStatus_COMPLETED

	// Mock UPDATE query - only title, supervisor_grade, status are provided
	// Args: title, supervisor_grade, status, updated_by, id
	mock.ExpectExec("UPDATE Final SET").
		WithArgs(
			newTitle,
			newSupervisorGrade,
			"completed", // status enum as string
			fixtures.UpdatedBy,
			fixtures.FinalID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetFinal
	rows := sqlmock.NewRows([]string{
		"id", "title", "supervisor_grade", "department_grade", "final_grade", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FinalID1,
		newTitle,
		newSupervisorGrade,
		fixtures.FinalDepartmentGrade1,
		fixtures.FinalGrade1,
		"completed",
		fixtures.FinalNotes1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Final WHERE id").
		WithArgs(fixtures.FinalID1).
		WillReturnRows(rows)

	req := &pb.UpdateFinalRequest{
		Id:              fixtures.FinalID1,
		Title:           &newTitle,
		SupervisorGrade: &newSupervisorGrade,
		Status:          &newStatus,
		// Other fields are nil - handler won't include them in UPDATE
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateFinal(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Final)
	assert.Equal(t, newTitle, resp.Final.Title)
	assert.Equal(t, newSupervisorGrade, resp.Final.SupervisorGrade)
	assert.Equal(t, pb.FinalStatus_COMPLETED, resp.Final.Status)
}

// TestUpdateFinal_NotFound tests updating non-existent final
func TestUpdateFinal_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Title"

	mock.ExpectExec("UPDATE Final SET").
		WithArgs(
			newTitle,
			fixtures.UpdatedBy,
			fixtures.FinalID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Final WHERE id").
		WithArgs(fixtures.FinalID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateFinalRequest{
		Id:        fixtures.FinalID1,
		Title:     &newTitle,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateFinal_NoFieldsToUpdate tests update with no fields
func TestUpdateFinal_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateFinalRequest{
		Id:        fixtures.FinalID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteFinal_Success tests successful final deletion
func TestDeleteFinal_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Final WHERE id").
		WithArgs(fixtures.FinalID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteFinalRequest{
		Id: fixtures.FinalID1,
	}

	resp, err := h.DeleteFinal(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteFinal_NotFound tests deleting non-existent final
func TestDeleteFinal_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Final WHERE id").
		WithArgs(fixtures.FinalID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteFinalRequest{
		Id: fixtures.FinalID1,
	}

	resp, err := h.DeleteFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteFinal_MissingID tests deletion with missing ID
func TestDeleteFinal_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteFinalRequest{
		Id: "",
	}

	resp, err := h.DeleteFinal(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListFinals_Success tests successful finals listing
func TestListFinals_Success(t *testing.T) {
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
		"id", "title", "supervisor_grade", "department_grade", "final_grade", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.FinalID1,
			fixtures.FinalTitle1,
			fixtures.FinalSupervisorGrade1,
			fixtures.FinalDepartmentGrade1,
			fixtures.FinalGrade1,
			"passed",
			fixtures.FinalNotes1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.FinalID2,
			fixtures.FinalTitle2,
			fixtures.FinalSupervisorGrade2,
			fixtures.FinalDepartmentGrade2,
			fixtures.FinalGrade2,
			"completed",
			fixtures.FinalNotes2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Final").
		WillReturnRows(rows)

	req := &pb.ListFinalsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListFinals(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.Finals, 2)
	assert.Equal(t, fixtures.FinalTitle1, resp.Finals[0].Title)
	assert.Equal(t, pb.FinalStatus_PASSED, resp.Finals[0].Status)
}

// TestListFinals_Empty tests listing with no results
func TestListFinals_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "title", "supervisor_grade", "department_grade", "final_grade", "status", "notes",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Final").
		WillReturnRows(rows)

	req := &pb.ListFinalsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListFinals(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.Finals)
}
