package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"thaily/src/service/file/handler"
	"thaily/src/service/file/tests/fixtures"

	common "thaily/proto/common"
	pb "thaily/proto/file"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestCreateFile_Success tests successful file creation
func TestCreateFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO File").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.FileTitle1,
			fixtures.FileName1,
			"file_pending", // status enum as string
			"topic",        // table enum as string
			fixtures.FileOption1,
			fixtures.FileTableID1,
			fixtures.CreatedBy,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetFile
	rows := sqlmock.NewRows([]string{
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FileID1,
		fixtures.FileTitle1,
		fixtures.FileName1,
		"file_pending",
		"topic",
		fixtures.FileOption1,
		fixtures.FileTableID1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateFileRequest{
		Title:     fixtures.FileTitle1,
		File:      fixtures.FileName1,
		Status:    pb.FileStatus_FILE_PENDING,
		Table:     pb.TableType_TOPIC,
		Option:    fixtures.FileOption1,
		TableId:   fixtures.FileTableID1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.File)
	assert.Equal(t, fixtures.FileTitle1, resp.File.Title)
	assert.Equal(t, pb.FileStatus_FILE_PENDING, resp.File.Status)
	assert.Equal(t, pb.TableType_TOPIC, resp.File.Table)
}

// TestCreateFile_MissingTitle tests creation with missing title
func TestCreateFile_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateFileRequest{
		Title:     "",
		File:      fixtures.FileName1,
		Status:    pb.FileStatus_FILE_PENDING,
		Table:     pb.TableType_TOPIC,
		Option:    fixtures.FileOption1,
		TableId:   fixtures.FileTableID1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateFile_MissingFile tests creation with missing file
func TestCreateFile_MissingFile(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateFileRequest{
		Title:     fixtures.FileTitle1,
		File:      "",
		Status:    pb.FileStatus_FILE_PENDING,
		Table:     pb.TableType_TOPIC,
		Option:    fixtures.FileOption1,
		TableId:   fixtures.FileTableID1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateFile_MissingOption tests creation with missing option
func TestCreateFile_MissingOption(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateFileRequest{
		Title:     fixtures.FileTitle1,
		File:      fixtures.FileName1,
		Status:    pb.FileStatus_FILE_PENDING,
		Table:     pb.TableType_TOPIC,
		Option:    "",
		TableId:   fixtures.FileTableID1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateFile_MissingTableId tests creation with missing table_id
func TestCreateFile_MissingTableId(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateFileRequest{
		Title:     fixtures.FileTitle1,
		File:      fixtures.FileName1,
		Status:    pb.FileStatus_FILE_PENDING,
		Table:     pb.TableType_TOPIC,
		Option:    fixtures.FileOption1,
		TableId:   "",
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestGetFile_Success tests successful file retrieval
func TestGetFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FileID1,
		fixtures.FileTitle1,
		fixtures.FileName1,
		"file_pending",
		"topic",
		fixtures.FileOption1,
		fixtures.FileTableID1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnRows(rows)

	req := &pb.GetFileRequest{
		Id: fixtures.FileID1,
	}

	resp, err := h.GetFile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.File)
	assert.Equal(t, fixtures.FileID1, resp.File.Id)
	assert.Equal(t, fixtures.FileTitle1, resp.File.Title)
	assert.Equal(t, pb.FileStatus_FILE_PENDING, resp.File.Status)
	assert.Equal(t, pb.TableType_TOPIC, resp.File.Table)
}

// TestGetFile_NotFound tests file not found
func TestGetFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetFileRequest{
		Id: fixtures.FileID1,
	}

	resp, err := h.GetFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetFile_MissingID tests getting file with missing ID
func TestGetFile_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetFileRequest{
		Id: "",
	}

	resp, err := h.GetFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateFile_Success tests successful file update
func TestUpdateFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Title"
	newStatus := pb.FileStatus_APPROVED

	// Mock UPDATE query - title and status provided
	mock.ExpectExec("UPDATE File SET").
		WithArgs(
			newTitle,
			"approved", // status enum as string
			fixtures.UpdatedBy,
			fixtures.FileID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetFile
	rows := sqlmock.NewRows([]string{
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FileID1,
		newTitle,
		fixtures.FileName1,
		"approved",
		"topic",
		fixtures.FileOption1,
		fixtures.FileTableID1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnRows(rows)

	req := &pb.UpdateFileRequest{
		Id:        fixtures.FileID1,
		Title:     &newTitle,
		Status:    &newStatus,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateFile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.File)
	assert.Equal(t, newTitle, resp.File.Title)
	assert.Equal(t, pb.FileStatus_APPROVED, resp.File.Status)
}

// TestUpdateFile_NotFound tests updating non-existent file
func TestUpdateFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Title"

	mock.ExpectExec("UPDATE File SET").
		WithArgs(
			newTitle,
			fixtures.UpdatedBy,
			fixtures.FileID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateFileRequest{
		Id:        fixtures.FileID1,
		Title:     &newTitle,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateFile_NoFieldsToUpdate tests update with no fields
func TestUpdateFile_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateFileRequest{
		Id:        fixtures.FileID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteFile_Success tests successful file deletion
func TestDeleteFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteFileRequest{
		Id: fixtures.FileID1,
	}

	resp, err := h.DeleteFile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteFile_NotFound tests deleting non-existent file
func TestDeleteFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteFileRequest{
		Id: fixtures.FileID1,
	}

	resp, err := h.DeleteFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteFile_MissingID tests deletion with missing ID
func TestDeleteFile_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteFileRequest{
		Id: "",
	}

	resp, err := h.DeleteFile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListFiles_Success tests successful files listing
func TestListFiles_Success(t *testing.T) {
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
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.FileID1,
			fixtures.FileTitle1,
			fixtures.FileName1,
			"file_pending",
			"topic",
			fixtures.FileOption1,
			fixtures.FileTableID1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.FileID2,
			fixtures.FileTitle2,
			fixtures.FileName2,
			"approved",
			"midterm",
			fixtures.FileOption2,
			fixtures.FileTableID2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM File").
		WillReturnRows(rows)

	req := &pb.ListFilesRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListFiles(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.Files, 2)
	assert.Equal(t, fixtures.FileTitle1, resp.Files[0].Title)
	assert.Equal(t, pb.FileStatus_FILE_PENDING, resp.Files[0].Status)
	assert.Equal(t, pb.TableType_TOPIC, resp.Files[0].Table)
	assert.Equal(t, fixtures.FileTitle2, resp.Files[1].Title)
	assert.Equal(t, pb.FileStatus_APPROVED, resp.Files[1].Status)
	assert.Equal(t, pb.TableType_MIDTERM, resp.Files[1].Table)
}

// TestListFiles_Empty tests listing with no results
func TestListFiles_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM File").
		WillReturnRows(rows)

	req := &pb.ListFilesRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListFiles(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.Files)
}

// TestCreateFile_WithDifferentEnums tests file creation with different enum values
func TestCreateFile_WithDifferentEnums(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Test with REJECTED status and FINAL table
	mock.ExpectExec("INSERT INTO File").
		WithArgs(
			sqlmock.AnyArg(),
			fixtures.FileTitle2,
			fixtures.FileName2,
			"rejected", // REJECTED status
			"final",    // FINAL table
			fixtures.FileOption2,
			fixtures.FileTableID2,
			fixtures.CreatedBy,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FileID2,
		fixtures.FileTitle2,
		fixtures.FileName2,
		"rejected",
		"final",
		fixtures.FileOption2,
		fixtures.FileTableID2,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateFileRequest{
		Title:     fixtures.FileTitle2,
		File:      fixtures.FileName2,
		Status:    pb.FileStatus_REJECTED,
		Table:     pb.TableType_FINAL,
		Option:    fixtures.FileOption2,
		TableId:   fixtures.FileTableID2,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateFile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.File)
	assert.Equal(t, pb.FileStatus_REJECTED, resp.File.Status)
	assert.Equal(t, pb.TableType_FINAL, resp.File.Table)
}

// TestUpdateFile_AllFields tests updating all updatable fields
func TestUpdateFile_AllFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Completely Updated Title"
	newFile := "newfile.pdf"
	newStatus := pb.FileStatus_APPROVED
	newTable := pb.TableType_ORDER
	newOption := "new_option"
	newTableId := "new-table-id"

	// Mock UPDATE query with all fields
	mock.ExpectExec("UPDATE File SET").
		WithArgs(
			newTitle,
			newFile,
			"approved",
			"order",
			newOption,
			newTableId,
			fixtures.UpdatedBy,
			fixtures.FileID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{
		"id", "title", "file", "status", "table", "option", "table_id",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.FileID1,
		newTitle,
		newFile,
		"approved",
		"order",
		newOption,
		newTableId,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM File WHERE id").
		WithArgs(fixtures.FileID1).
		WillReturnRows(rows)

	req := &pb.UpdateFileRequest{
		Id:        fixtures.FileID1,
		Title:     &newTitle,
		File:      &newFile,
		Status:    &newStatus,
		Table:     &newTable,
		Option:    &newOption,
		TableId:   &newTableId,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateFile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.File)
	assert.Equal(t, newTitle, resp.File.Title)
	assert.Equal(t, newFile, resp.File.File)
	assert.Equal(t, pb.FileStatus_APPROVED, resp.File.Status)
	assert.Equal(t, pb.TableType_ORDER, resp.File.Table)
	assert.Equal(t, newOption, resp.File.Option)
	assert.Equal(t, newTableId, resp.File.TableId)
}
