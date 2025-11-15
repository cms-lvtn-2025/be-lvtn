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

// ==================== CREATE ENROLLMENT TESTS ====================

func TestCreateEnrollment_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateEnrollmentRequest()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Enrollment").
		WithArgs(
			sqlmock.AnyArg(), // id
			req.Title,
			req.StudentCode,
			req.TopicCouncilCode,
			*req.FinalCode,
			"", // grade_review_code (nil -> "")
			*req.MidtermCode,
			req.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT query (GetEnrollment is called after insert)
	// Column order: id, title, student_code, topic_council_code, final_code, grade_review_code, midterm_code, created_at, updated_at, created_by, updated_by
	mock.ExpectQuery("SELECT (.+) FROM Enrollment WHERE id = ?").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "student_code", "topic_council_code",
			"final_code", "grade_review_code", "midterm_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			"test-id", req.Title, req.StudentCode, req.TopicCouncilCode,
			*req.FinalCode, "", *req.MidtermCode,
			time.Now(), time.Now(), req.CreatedBy, "",
		))

	// Execute
	resp, err := h.CreateEnrollment(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Enrollment)
	assert.Equal(t, req.Title, resp.Enrollment.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateEnrollment_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := &pb.CreateEnrollmentRequest{
		Title:            "",
		StudentCode:      "STU001",
		TopicCouncilCode: "TC001",
		CreatedBy:        "test-user",
	}

	// Execute
	resp, err := h.CreateEnrollment(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateEnrollment_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateEnrollmentRequest()

	// Mock INSERT with error
	mock.ExpectExec("INSERT INTO Enrollment").
		WillReturnError(sql.ErrConnDone)

	// Execute
	resp, err := h.CreateEnrollment(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GET ENROLLMENT TESTS ====================

func TestGetEnrollment_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	enrollment := fixtures.GetTestEnrollment()

	// Mock SELECT query
	mock.ExpectQuery("SELECT (.+) FROM Enrollment WHERE id = ?").
		WithArgs(enrollment.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "student_code", "topic_council_code",
			"final_code", "grade_review_code", "midterm_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			enrollment.Id, enrollment.Title, enrollment.StudentCode, enrollment.TopicCouncilCode,
			*enrollment.FinalCode, "", *enrollment.MidtermCode,
			enrollment.CreatedAt.AsTime(), enrollment.UpdatedAt.AsTime(),
			enrollment.CreatedBy, enrollment.UpdatedBy,
		))

	// Execute
	resp, err := h.GetEnrollment(context.Background(), &pb.GetEnrollmentRequest{Id: enrollment.Id})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Enrollment)
	assert.Equal(t, enrollment.Id, resp.Enrollment.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEnrollment_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock SELECT with no rows
	mock.ExpectQuery("SELECT (.+) FROM Enrollment WHERE id = ?").
		WithArgs(fixtures.TestEnrollmentID1).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.GetEnrollment(context.Background(), &pb.GetEnrollmentRequest{Id: fixtures.TestEnrollmentID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEnrollment_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.GetEnrollment(context.Background(), &pb.GetEnrollmentRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== UPDATE ENROLLMENT TESTS ====================

func TestUpdateEnrollment_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateEnrollmentRequest()

	// Mock UPDATE - midterm_code is nil, so it's not included in UPDATE
	// Only 6 fields + updated_by (7 args total, + id at end = 8)
	// But wait - handler only adds fields that are NOT nil
	// title, student_code, topic_council_code, final_code, grade_review_code, updated_by, id = 7 args
	mock.ExpectExec("UPDATE Enrollment SET").
		WithArgs(
			*req.Title,
			*req.StudentCode,
			*req.TopicCouncilCode,
			*req.FinalCode,
			*req.GradeReviewCode,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT for GetEnrollment (called after update)
	mock.ExpectQuery("SELECT (.+) FROM Enrollment WHERE id = ?").
		WithArgs(req.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "student_code", "topic_council_code",
			"final_code", "grade_review_code", "midterm_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			req.Id, *req.Title, *req.StudentCode, *req.TopicCouncilCode,
			*req.FinalCode, *req.GradeReviewCode, "",
			time.Now(), time.Now(), "test-user", req.UpdatedBy,
		))

	// Execute
	resp, err := h.UpdateEnrollment(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Enrollment)
	assert.Equal(t, *req.Title, resp.Enrollment.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateEnrollment_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateEnrollmentRequest()

	// Mock UPDATE - only non-nil fields
	mock.ExpectExec("UPDATE Enrollment SET").
		WithArgs(
			*req.Title,
			*req.StudentCode,
			*req.TopicCouncilCode,
			*req.FinalCode,
			*req.GradeReviewCode,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT that returns NotFound
	mock.ExpectQuery("SELECT (.+) FROM Enrollment WHERE id = ?").
		WithArgs(req.Id).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.UpdateEnrollment(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateEnrollment_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	title := "Updated"
	req := &pb.UpdateEnrollmentRequest{
		Id:        "",
		Title:     &title,
		UpdatedBy: "test-user",
	}

	// Execute
	resp, err := h.UpdateEnrollment(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== DELETE ENROLLMENT TESTS ====================

func TestDeleteEnrollment_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE
	mock.ExpectExec("DELETE FROM Enrollment WHERE id = ?").
		WithArgs(fixtures.TestEnrollmentID1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	resp, err := h.DeleteEnrollment(context.Background(), &pb.DeleteEnrollmentRequest{Id: fixtures.TestEnrollmentID1})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEnrollment_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE with no rows
	mock.ExpectExec("DELETE FROM Enrollment WHERE id = ?").
		WithArgs(fixtures.TestEnrollmentID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	resp, err := h.DeleteEnrollment(context.Background(), &pb.DeleteEnrollmentRequest{Id: fixtures.TestEnrollmentID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEnrollment_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.DeleteEnrollment(context.Background(), &pb.DeleteEnrollmentRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== LIST ENROLLMENTS TESTS ====================

func TestListEnrollments_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	enrollment1 := fixtures.GetTestEnrollment()
	enrollment2 := fixtures.GetTestEnrollment2()

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Enrollment").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock SELECT
	mock.ExpectQuery("SELECT (.+) FROM Enrollment").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "student_code", "topic_council_code",
			"final_code", "grade_review_code", "midterm_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}).
			AddRow(
				enrollment1.Id, enrollment1.Title, enrollment1.StudentCode, enrollment1.TopicCouncilCode,
				*enrollment1.FinalCode, "", *enrollment1.MidtermCode,
				enrollment1.CreatedAt.AsTime(), enrollment1.UpdatedAt.AsTime(),
				enrollment1.CreatedBy, enrollment1.UpdatedBy,
			).
			AddRow(
				enrollment2.Id, enrollment2.Title, enrollment2.StudentCode, enrollment2.TopicCouncilCode,
				"", *enrollment2.GradeReviewCode, "",
				enrollment2.CreatedAt.AsTime(), enrollment2.UpdatedAt.AsTime(),
				enrollment2.CreatedBy, enrollment2.UpdatedBy,
			))

	// Execute
	req := fixtures.GetTestListEnrollmentsRequest()
	resp, err := h.ListEnrollments(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Enrollments, 2)
	assert.Equal(t, int32(2), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListEnrollments_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Enrollment").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock SELECT
	mock.ExpectQuery("SELECT (.+) FROM Enrollment").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "student_code", "topic_council_code",
			"final_code", "grade_review_code", "midterm_code",
			"created_at", "updated_at", "created_by", "updated_by",
		}))

	// Execute
	req := fixtures.GetTestListEnrollmentsRequest()
	resp, err := h.ListEnrollments(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Enrollments, 0)
	assert.Equal(t, int32(0), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListEnrollments_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT with error
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Enrollment").
		WillReturnError(sql.ErrConnDone)

	// Execute
	req := fixtures.GetTestListEnrollmentsRequest()
	resp, err := h.ListEnrollments(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}
