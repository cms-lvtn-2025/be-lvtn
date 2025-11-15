package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"thaily/src/service/thesis/handler"
	"thaily/src/service/thesis/tests/fixtures"

	pb "thaily/proto/thesis"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ==================== CREATE TOPIC TESTS ====================

func TestCreateTopic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateTopicRequest()

	// Mock INSERT query - status enum is converted to string "submit"
	mock.ExpectExec("INSERT INTO Topic").
		WithArgs(
			sqlmock.AnyArg(), // id
			req.Title,
			req.MajorCode,
			req.SemesterCode,
			"submit", // Status enum converted to string
			*req.PercentStage_1,
			*req.PercentStage_2,
			req.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT query (GetTopic is called after insert)
	// Column order: id, title, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_at, updated_at, created_by, updated_by
	mock.ExpectQuery("SELECT (.+) FROM Topic WHERE id = ?").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code", "status",
			"percent_stage_1", "percent_stage_2", "created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			"test-id", req.Title, req.MajorCode, req.SemesterCode, "submit",
			*req.PercentStage_1, *req.PercentStage_2, time.Now(), time.Now(), req.CreatedBy, "",
		))

	// Execute
	resp, err := h.CreateTopic(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Topic)
	assert.Equal(t, req.Title, resp.Topic.Title)
	assert.Equal(t, pb.TopicStatus_SUBMIT, resp.Topic.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTopic_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	percent1 := int32(50)
	percent2 := int32(50)
	req := &pb.CreateTopicRequest{
		Title:          "",
		MajorCode:      "CS",
		SemesterCode:   "2024-1",
		Status:         pb.TopicStatus_SUBMIT,
		PercentStage_1: &percent1,
		PercentStage_2: &percent2,
		CreatedBy:      "test-user",
	}

	// Execute
	resp, err := h.CreateTopic(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateTopic_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateTopicRequest()

	// Mock INSERT with error
	mock.ExpectExec("INSERT INTO Topic").
		WillReturnError(sql.ErrConnDone)

	// Execute
	resp, err := h.CreateTopic(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GET TOPIC TESTS ====================

func TestGetTopic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	topic := fixtures.GetTestTopic()

	// Mock SELECT query - status is stored as string "submit" in DB
	mock.ExpectQuery("SELECT (.+) FROM Topic WHERE id = ?").
		WithArgs(topic.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code", "status",
			"percent_stage_1", "percent_stage_2", "created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			topic.Id, topic.Title, topic.MajorCode, topic.SemesterCode, "submit",
			*topic.PercentStage_1, *topic.PercentStage_2,
			topic.CreatedAt.AsTime(), topic.UpdatedAt.AsTime(),
			topic.CreatedBy, topic.UpdatedBy,
		))

	// Execute
	resp, err := h.GetTopic(context.Background(), &pb.GetTopicRequest{Id: topic.Id})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Topic)
	assert.Equal(t, topic.Id, resp.Topic.Id)
	assert.Equal(t, pb.TopicStatus_SUBMIT, resp.Topic.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTopic_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock SELECT with no rows
	mock.ExpectQuery("SELECT (.+) FROM Topic WHERE id = ?").
		WithArgs(fixtures.TestTopicID1).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.GetTopic(context.Background(), &pb.GetTopicRequest{Id: fixtures.TestTopicID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTopic_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.GetTopic(context.Background(), &pb.GetTopicRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== UPDATE TOPIC TESTS ====================

func TestUpdateTopic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateTopicRequest()

	// Mock UPDATE - status enum is converted to "approved_1"
	mock.ExpectExec("UPDATE Topic SET").
		WithArgs(
			*req.Title,
			*req.MajorCode,
			*req.SemesterCode,
			"approved_1", // Status enum converted
			*req.PercentStage_1,
			*req.PercentStage_2,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT for GetTopic (called after update)
	mock.ExpectQuery("SELECT (.+) FROM Topic WHERE id = ?").
		WithArgs(req.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code", "status",
			"percent_stage_1", "percent_stage_2", "created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			req.Id, *req.Title, *req.MajorCode, *req.SemesterCode, "approved_1",
			*req.PercentStage_1, *req.PercentStage_2,
			time.Now(), time.Now(), "test-user", req.UpdatedBy,
		))

	// Execute
	resp, err := h.UpdateTopic(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Topic)
	assert.Equal(t, *req.Title, resp.Topic.Title)
	assert.Equal(t, pb.TopicStatus_APPROVED_1, resp.Topic.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTopic_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateTopicRequest()

	// Mock UPDATE
	mock.ExpectExec("UPDATE Topic SET").
		WithArgs(
			*req.Title,
			*req.MajorCode,
			*req.SemesterCode,
			"approved_1",
			*req.PercentStage_1,
			*req.PercentStage_2,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT that returns NotFound
	mock.ExpectQuery("SELECT (.+) FROM Topic WHERE id = ?").
		WithArgs(req.Id).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.UpdateTopic(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTopic_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	title := "Updated"
	req := &pb.UpdateTopicRequest{
		Id:        "",
		Title:     &title,
		UpdatedBy: "test-user",
	}

	// Execute
	resp, err := h.UpdateTopic(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== DELETE TOPIC TESTS ====================

func TestDeleteTopic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE
	mock.ExpectExec("DELETE FROM Topic WHERE id = ?").
		WithArgs(fixtures.TestTopicID1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	resp, err := h.DeleteTopic(context.Background(), &pb.DeleteTopicRequest{Id: fixtures.TestTopicID1})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTopic_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE with no rows
	mock.ExpectExec("DELETE FROM Topic WHERE id = ?").
		WithArgs(fixtures.TestTopicID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	resp, err := h.DeleteTopic(context.Background(), &pb.DeleteTopicRequest{Id: fixtures.TestTopicID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTopic_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.DeleteTopic(context.Background(), &pb.DeleteTopicRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== LIST TOPICS TESTS ====================

func TestListTopics_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	topic1 := fixtures.GetTestTopic()
	topic2 := fixtures.GetTestTopic2()

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Topic").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock SELECT - statuses stored as strings in DB
	mock.ExpectQuery("SELECT (.+) FROM Topic").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code", "status",
			"percent_stage_1", "percent_stage_2", "created_at", "updated_at", "created_by", "updated_by",
		}).
			AddRow(
				topic1.Id, topic1.Title, topic1.MajorCode, topic1.SemesterCode, "submit",
				*topic1.PercentStage_1, *topic1.PercentStage_2,
				topic1.CreatedAt.AsTime(), topic1.UpdatedAt.AsTime(),
				topic1.CreatedBy, topic1.UpdatedBy,
			).
			AddRow(
				topic2.Id, topic2.Title, topic2.MajorCode, topic2.SemesterCode, "in_progress",
				*topic2.PercentStage_1, *topic2.PercentStage_2,
				topic2.CreatedAt.AsTime(), topic2.UpdatedAt.AsTime(),
				topic2.CreatedBy, topic2.UpdatedBy,
			))

	// Execute
	req := fixtures.GetTestListTopicsRequest()
	resp, err := h.ListTopics(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Topics, 2)
	assert.Equal(t, int32(2), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListTopics_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Topic").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock SELECT
	mock.ExpectQuery("SELECT (.+) FROM Topic").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "major_code", "semester_code", "status",
			"percent_stage_1", "percent_stage_2", "created_at", "updated_at", "created_by", "updated_by",
		}))

	// Execute
	req := fixtures.GetTestListTopicsRequest()
	resp, err := h.ListTopics(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Topics, 0)
	assert.Equal(t, int32(0), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListTopics_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT with error
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Topic").
		WillReturnError(sql.ErrConnDone)

	// Execute
	req := fixtures.GetTestListTopicsRequest()
	resp, err := h.ListTopics(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}
