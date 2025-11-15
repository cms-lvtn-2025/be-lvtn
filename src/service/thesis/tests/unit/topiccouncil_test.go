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

// TestCreateTopicCouncil_Success tests successful topic council creation
func TestCreateTopicCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO TopicCouncil").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.TopicCouncilTitle1,
			"stage_dacn", // stage enum as string
			fixtures.TopicCouncilTopicCode1,
			fixtures.TopicCouncilCouncilCode1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetTopicCouncil
	rows := sqlmock.NewRows([]string{
		"id", "title", "stage", "topic_code", "council_code", "time_start", "time_end",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TopicCouncilID1,
		fixtures.TopicCouncilTitle1,
		"stage_dacn",
		fixtures.TopicCouncilTopicCode1,
		fixtures.TopicCouncilCouncilCode1,
		time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Topic_council WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateTopicCouncilRequest{
		Title:       fixtures.TopicCouncilTitle1,
		Stage:       pb.TopicStage_STAGE_DACN,
		TopicCode:   fixtures.TopicCouncilTopicCode1,
		CouncilCode: &fixtures.TopicCouncilCouncilCode1,
		CreatedBy:   fixtures.CreatedBy,
	}

	resp, err := h.CreateTopicCouncil(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.TopicCouncil)
	assert.Equal(t, fixtures.TopicCouncilTitle1, resp.TopicCouncil.Title)
}

// TestCreateTopicCouncil_MissingTitle tests creation with missing title
func TestCreateTopicCouncil_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateTopicCouncilRequest{
		Title:     "",
		TopicCode: fixtures.TopicCouncilTopicCode1,
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateTopicCouncil_MissingTopicCode tests creation with missing topic code
func TestCreateTopicCouncil_MissingTopicCode(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateTopicCouncilRequest{
		Title:     fixtures.TopicCouncilTitle1,
		TopicCode: "",
		CreatedBy: fixtures.CreatedBy,
	}

	resp, err := h.CreateTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestGetTopicCouncil_Success tests successful topic council retrieval
func TestGetTopicCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "title", "stage", "topic_code", "council_code", "time_start", "time_end",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TopicCouncilID1,
		fixtures.TopicCouncilTitle1,
		"stage_dacn",
		fixtures.TopicCouncilTopicCode1,
		fixtures.TopicCouncilCouncilCode1,
		time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Topic_council WHERE id").
		WithArgs(fixtures.TopicCouncilID1).
		WillReturnRows(rows)

	req := &pb.GetTopicCouncilRequest{
		Id: fixtures.TopicCouncilID1,
	}

	resp, err := h.GetTopicCouncil(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.TopicCouncil)
	assert.Equal(t, fixtures.TopicCouncilID1, resp.TopicCouncil.Id)
	assert.Equal(t, fixtures.TopicCouncilTitle1, resp.TopicCouncil.Title)
	assert.Equal(t, pb.TopicStage_STAGE_DACN, resp.TopicCouncil.Stage)
}

// TestGetTopicCouncil_NotFound tests topic council not found
func TestGetTopicCouncil_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Topic_council WHERE id").
		WithArgs(fixtures.TopicCouncilID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetTopicCouncilRequest{
		Id: fixtures.TopicCouncilID1,
	}

	resp, err := h.GetTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetTopicCouncil_MissingID tests getting topic council with missing ID
func TestGetTopicCouncil_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetTopicCouncilRequest{
		Id: "",
	}

	resp, err := h.GetTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateTopicCouncil_Success tests successful topic council update
// Note: Handler uses dynamic UPDATE - only includes non-nil fields
func TestUpdateTopicCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Topic Council Title"
	newStage := pb.TopicStage_STAGE_LVTN

	// Mock UPDATE query - only title and stage are provided
	// Args: title, stage, updated_by, id
	mock.ExpectExec("UPDATE TopicCouncil SET").
		WithArgs(
			newTitle,
			"stage_lvtn", // stage enum as string
			fixtures.UpdatedBy,
			fixtures.TopicCouncilID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetTopicCouncil
	rows := sqlmock.NewRows([]string{
		"id", "title", "stage", "topic_code", "council_code", "time_start", "time_end",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TopicCouncilID1,
		newTitle,
		"stage_lvtn",
		fixtures.TopicCouncilTopicCode1,
		fixtures.TopicCouncilCouncilCode1,
		time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Topic_council WHERE id").
		WithArgs(fixtures.TopicCouncilID1).
		WillReturnRows(rows)

	req := &pb.UpdateTopicCouncilRequest{
		Id:    fixtures.TopicCouncilID1,
		Title: &newTitle,
		Stage: &newStage,
		// TopicCode and CouncilCode are nil - handler won't include them in UPDATE
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateTopicCouncil(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.TopicCouncil)
	assert.Equal(t, newTitle, resp.TopicCouncil.Title)
	assert.Equal(t, pb.TopicStage_STAGE_LVTN, resp.TopicCouncil.Stage)
}

// TestUpdateTopicCouncil_NotFound tests updating non-existent topic council
func TestUpdateTopicCouncil_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTitle := "Updated Title"

	mock.ExpectExec("UPDATE TopicCouncil SET").
		WithArgs(
			newTitle,
			fixtures.UpdatedBy,
			fixtures.TopicCouncilID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Topic_council WHERE id").
		WithArgs(fixtures.TopicCouncilID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateTopicCouncilRequest{
		Id:        fixtures.TopicCouncilID1,
		Title:     &newTitle,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateTopicCouncil_NoFieldsToUpdate tests update with no fields
func TestUpdateTopicCouncil_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateTopicCouncilRequest{
		Id:        fixtures.TopicCouncilID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteTopicCouncil_Success tests successful topic council deletion
func TestDeleteTopicCouncil_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Topic_council WHERE id").
		WithArgs(fixtures.TopicCouncilID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteTopicCouncilRequest{
		Id: fixtures.TopicCouncilID1,
	}

	resp, err := h.DeleteTopicCouncil(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteTopicCouncil_NotFound tests deleting non-existent topic council
func TestDeleteTopicCouncil_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Topic_council WHERE id").
		WithArgs(fixtures.TopicCouncilID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteTopicCouncilRequest{
		Id: fixtures.TopicCouncilID1,
	}

	resp, err := h.DeleteTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteTopicCouncil_MissingID tests deletion with missing ID
func TestDeleteTopicCouncil_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteTopicCouncilRequest{
		Id: "",
	}

	resp, err := h.DeleteTopicCouncil(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListTopicCouncils_Success tests successful topic councils listing
func TestListTopicCouncils_Success(t *testing.T) {
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
		"id", "title", "stage", "topic_code", "council_code", "time_start", "time_end",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.TopicCouncilID1,
			fixtures.TopicCouncilTitle1,
			"stage_dacn",
			fixtures.TopicCouncilTopicCode1,
			fixtures.TopicCouncilCouncilCode1,
			time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC),
			time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.TopicCouncilID2,
			fixtures.TopicCouncilTitle2,
			"stage_lvtn",
			fixtures.TopicCouncilTopicCode2,
			fixtures.TopicCouncilCouncilCode2,
			time.Date(2024, 6, 2, 9, 0, 0, 0, time.UTC),
			time.Date(2024, 6, 2, 12, 0, 0, 0, time.UTC),
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Topic_council").
		WillReturnRows(rows)

	req := &pb.ListTopicCouncilsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListTopicCouncils(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.TopicCouncils, 2)
	assert.Equal(t, fixtures.TopicCouncilTitle1, resp.TopicCouncils[0].Title)
	assert.Equal(t, pb.TopicStage_STAGE_DACN, resp.TopicCouncils[0].Stage)
}

// TestListTopicCouncils_Empty tests listing with no results
func TestListTopicCouncils_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "title", "stage", "topic_code", "council_code", "time_start", "time_end",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Topic_council").
		WillReturnRows(rows)

	req := &pb.ListTopicCouncilsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListTopicCouncils(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.TopicCouncils)
}
