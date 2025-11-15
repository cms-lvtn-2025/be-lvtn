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

// TestCreateTopicCouncilSupervisor_Success tests successful topic council supervisor creation
func TestCreateTopicCouncilSupervisor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Topic_council_supervisor").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.TopicCouncilSupervisorTeacherCode1,
			fixtures.TopicCouncilSupervisorCouncilCode1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetTopicCouncilSupervisor
	rows := sqlmock.NewRows([]string{
		"id", "teacher_supervisor_code", "topic_council_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TopicCouncilSupervisorID1,
		fixtures.TopicCouncilSupervisorTeacherCode1,
		fixtures.TopicCouncilSupervisorCouncilCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateTopicCouncilSupervisorRequest{
		TeacherSupervisorCode: fixtures.TopicCouncilSupervisorTeacherCode1,
		TopicCouncilCode:      fixtures.TopicCouncilSupervisorCouncilCode1,
		CreatedBy:             fixtures.CreatedBy,
	}

	resp, err := h.CreateTopicCouncilSupervisor(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.TopicCouncilSupervisor)
	assert.Equal(t, fixtures.TopicCouncilSupervisorTeacherCode1, resp.TopicCouncilSupervisor.TeacherSupervisorCode)
}

// TestCreateTopicCouncilSupervisor_MissingTeacherCode tests creation with missing teacher code
func TestCreateTopicCouncilSupervisor_MissingTeacherCode(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateTopicCouncilSupervisorRequest{
		TeacherSupervisorCode: "",
		TopicCouncilCode:      fixtures.TopicCouncilSupervisorCouncilCode1,
		CreatedBy:             fixtures.CreatedBy,
	}

	resp, err := h.CreateTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateTopicCouncilSupervisor_MissingTopicCouncilCode tests creation with missing topic council code
func TestCreateTopicCouncilSupervisor_MissingTopicCouncilCode(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateTopicCouncilSupervisorRequest{
		TeacherSupervisorCode: fixtures.TopicCouncilSupervisorTeacherCode1,
		TopicCouncilCode:      "",
		CreatedBy:             fixtures.CreatedBy,
	}

	resp, err := h.CreateTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestGetTopicCouncilSupervisor_Success tests successful topic council supervisor retrieval
func TestGetTopicCouncilSupervisor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "teacher_supervisor_code", "topic_council_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TopicCouncilSupervisorID1,
		fixtures.TopicCouncilSupervisorTeacherCode1,
		fixtures.TopicCouncilSupervisorCouncilCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor WHERE id").
		WithArgs(fixtures.TopicCouncilSupervisorID1).
		WillReturnRows(rows)

	req := &pb.GetTopicCouncilSupervisorRequest{
		Id: fixtures.TopicCouncilSupervisorID1,
	}

	resp, err := h.GetTopicCouncilSupervisor(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.TopicCouncilSupervisor)
	assert.Equal(t, fixtures.TopicCouncilSupervisorID1, resp.TopicCouncilSupervisor.Id)
	assert.Equal(t, fixtures.TopicCouncilSupervisorTeacherCode1, resp.TopicCouncilSupervisor.TeacherSupervisorCode)
}

// TestGetTopicCouncilSupervisor_NotFound tests topic council supervisor not found
func TestGetTopicCouncilSupervisor_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor WHERE id").
		WithArgs(fixtures.TopicCouncilSupervisorID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetTopicCouncilSupervisorRequest{
		Id: fixtures.TopicCouncilSupervisorID1,
	}

	resp, err := h.GetTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetTopicCouncilSupervisor_MissingID tests getting topic council supervisor with missing ID
func TestGetTopicCouncilSupervisor_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetTopicCouncilSupervisorRequest{
		Id: "",
	}

	resp, err := h.GetTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateTopicCouncilSupervisor_Success tests successful topic council supervisor update
func TestUpdateTopicCouncilSupervisor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTeacherCode := "TEACHER999"
	newCouncilCode := "COUNCIL999"

	// Mock UPDATE query
	// Args: teacher_supervisor_code, topic_council_code, updated_by, id
	mock.ExpectExec("UPDATE TopicCouncilSupervisor SET").
		WithArgs(
			newTeacherCode,
			newCouncilCode,
			fixtures.UpdatedBy,
			fixtures.TopicCouncilSupervisorID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetTopicCouncilSupervisor
	rows := sqlmock.NewRows([]string{
		"id", "teacher_supervisor_code", "topic_council_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TopicCouncilSupervisorID1,
		newTeacherCode,
		newCouncilCode,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor WHERE id").
		WithArgs(fixtures.TopicCouncilSupervisorID1).
		WillReturnRows(rows)

	req := &pb.UpdateTopicCouncilSupervisorRequest{
		Id:                    fixtures.TopicCouncilSupervisorID1,
		TeacherSupervisorCode: &newTeacherCode,
		TopicCouncilCode:      &newCouncilCode,
		UpdatedBy:             fixtures.UpdatedBy,
	}

	resp, err := h.UpdateTopicCouncilSupervisor(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.TopicCouncilSupervisor)
	assert.Equal(t, newTeacherCode, resp.TopicCouncilSupervisor.TeacherSupervisorCode)
	assert.Equal(t, newCouncilCode, resp.TopicCouncilSupervisor.TopicCouncilCode)
}

// TestUpdateTopicCouncilSupervisor_NotFound tests updating non-existent topic council supervisor
func TestUpdateTopicCouncilSupervisor_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newTeacherCode := "TEACHER999"

	mock.ExpectExec("UPDATE TopicCouncilSupervisor SET").
		WithArgs(
			newTeacherCode,
			fixtures.UpdatedBy,
			fixtures.TopicCouncilSupervisorID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor WHERE id").
		WithArgs(fixtures.TopicCouncilSupervisorID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateTopicCouncilSupervisorRequest{
		Id:                    fixtures.TopicCouncilSupervisorID1,
		TeacherSupervisorCode: &newTeacherCode,
		UpdatedBy:             fixtures.UpdatedBy,
	}

	resp, err := h.UpdateTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateTopicCouncilSupervisor_NoFieldsToUpdate tests update with no fields
func TestUpdateTopicCouncilSupervisor_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateTopicCouncilSupervisorRequest{
		Id:        fixtures.TopicCouncilSupervisorID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteTopicCouncilSupervisor_Success tests successful topic council supervisor deletion
func TestDeleteTopicCouncilSupervisor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Topic_council_supervisor WHERE id").
		WithArgs(fixtures.TopicCouncilSupervisorID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteTopicCouncilSupervisorRequest{
		Id: fixtures.TopicCouncilSupervisorID1,
	}

	resp, err := h.DeleteTopicCouncilSupervisor(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteTopicCouncilSupervisor_NotFound tests deleting non-existent topic council supervisor
func TestDeleteTopicCouncilSupervisor_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Topic_council_supervisor WHERE id").
		WithArgs(fixtures.TopicCouncilSupervisorID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteTopicCouncilSupervisorRequest{
		Id: fixtures.TopicCouncilSupervisorID1,
	}

	resp, err := h.DeleteTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteTopicCouncilSupervisor_MissingID tests deletion with missing ID
func TestDeleteTopicCouncilSupervisor_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteTopicCouncilSupervisorRequest{
		Id: "",
	}

	resp, err := h.DeleteTopicCouncilSupervisor(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListTopicCouncilSupervisors_Success tests successful topic council supervisors listing
func TestListTopicCouncilSupervisors_Success(t *testing.T) {
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
		"id", "teacher_supervisor_code", "topic_council_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.TopicCouncilSupervisorID1,
			fixtures.TopicCouncilSupervisorTeacherCode1,
			fixtures.TopicCouncilSupervisorCouncilCode1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.TopicCouncilSupervisorID2,
			fixtures.TopicCouncilSupervisorTeacherCode2,
			fixtures.TopicCouncilSupervisorCouncilCode2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor").
		WillReturnRows(rows)

	req := &pb.ListTopicCouncilSupervisorsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListTopicCouncilSupervisors(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.TopicCouncilSupervisors, 2)
	assert.Equal(t, fixtures.TopicCouncilSupervisorTeacherCode1, resp.TopicCouncilSupervisors[0].TeacherSupervisorCode)
}

// TestListTopicCouncilSupervisors_Empty tests listing with no results
func TestListTopicCouncilSupervisors_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "teacher_supervisor_code", "topic_council_code",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Topic_council_supervisor").
		WillReturnRows(rows)

	req := &pb.ListTopicCouncilSupervisorsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListTopicCouncilSupervisors(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.TopicCouncilSupervisors)
}
