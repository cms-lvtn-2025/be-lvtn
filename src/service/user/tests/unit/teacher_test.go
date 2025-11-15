package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"thaily/src/service/user/handler"
	"thaily/src/service/user/tests/fixtures"

	common "thaily/proto/common"
	pb "thaily/proto/user"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestCreateTeacher_Success tests successful teacher creation
func TestCreateTeacher_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Teacher").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.TeacherEmail1,
			fixtures.TeacherUsername1,
			"male", // gender enum as string
			fixtures.TeacherMajorCode1,
			fixtures.TeacherSemesterCode1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetTeacher
	rows := sqlmock.NewRows([]string{
		"id", "email", "username", "gender", "major_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TeacherID1,
		fixtures.TeacherEmail1,
		fixtures.TeacherUsername1,
		"male",
		fixtures.TeacherMajorCode1,
		fixtures.TeacherSemesterCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Teacher WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateTeacherRequest{
		Email:        fixtures.TeacherEmail1,
		Username:     fixtures.TeacherUsername1,
		Gender:       pb.Gender_MALE,
		MajorCode:    fixtures.TeacherMajorCode1,
		SemesterCode: fixtures.TeacherSemesterCode1,
		CreatedBy:    fixtures.CreatedBy,
	}

	resp, err := h.CreateTeacher(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Teacher)
	assert.Equal(t, fixtures.TeacherEmail1, resp.Teacher.Email)
}

// TestCreateTeacher_MissingEmail tests creation with missing email
func TestCreateTeacher_MissingEmail(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateTeacherRequest{
		Email:        "",
		Username:     fixtures.TeacherUsername1,
		Gender:       pb.Gender_MALE,
		MajorCode:    fixtures.TeacherMajorCode1,
		SemesterCode: fixtures.TeacherSemesterCode1,
		CreatedBy:    fixtures.CreatedBy,
	}

	resp, err := h.CreateTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateTeacher_MissingUsername tests creation with missing username
func TestCreateTeacher_MissingUsername(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateTeacherRequest{
		Email:        fixtures.TeacherEmail1,
		Username:     "",
		Gender:       pb.Gender_MALE,
		MajorCode:    fixtures.TeacherMajorCode1,
		SemesterCode: fixtures.TeacherSemesterCode1,
		CreatedBy:    fixtures.CreatedBy,
	}

	resp, err := h.CreateTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestGetTeacher_Success tests successful teacher retrieval
func TestGetTeacher_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "email", "username", "gender", "major_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TeacherID1,
		fixtures.TeacherEmail1,
		fixtures.TeacherUsername1,
		"male",
		fixtures.TeacherMajorCode1,
		fixtures.TeacherSemesterCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Teacher WHERE id").
		WithArgs(fixtures.TeacherID1).
		WillReturnRows(rows)

	req := &pb.GetTeacherRequest{
		Id: fixtures.TeacherID1,
	}

	resp, err := h.GetTeacher(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Teacher)
	assert.Equal(t, fixtures.TeacherID1, resp.Teacher.Id)
	assert.Equal(t, fixtures.TeacherEmail1, resp.Teacher.Email)
	assert.Equal(t, pb.Gender_MALE, resp.Teacher.Gender)
}

// TestGetTeacher_NotFound tests teacher not found
func TestGetTeacher_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Teacher WHERE id").
		WithArgs(fixtures.TeacherID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetTeacherRequest{
		Id: fixtures.TeacherID1,
	}

	resp, err := h.GetTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetTeacher_MissingID tests getting teacher with missing ID
func TestGetTeacher_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetTeacherRequest{
		Id: "",
	}

	resp, err := h.GetTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateTeacher_Success tests successful teacher update
func TestUpdateTeacher_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newEmail := "newemail@example.com"
	newUsername := "new_username"
	newGender := pb.Gender_FEMALE

	// Mock UPDATE query - email, username, gender provided
	mock.ExpectExec("UPDATE Teacher SET").
		WithArgs(
			newEmail,
			newUsername,
			"female", // gender enum as string
			fixtures.UpdatedBy,
			fixtures.TeacherID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetTeacher
	rows := sqlmock.NewRows([]string{
		"id", "email", "username", "gender", "major_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.TeacherID1,
		newEmail,
		newUsername,
		"female",
		fixtures.TeacherMajorCode1,
		fixtures.TeacherSemesterCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Teacher WHERE id").
		WithArgs(fixtures.TeacherID1).
		WillReturnRows(rows)

	req := &pb.UpdateTeacherRequest{
		Id:        fixtures.TeacherID1,
		Email:     &newEmail,
		Username:  &newUsername,
		Gender:    &newGender,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateTeacher(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Teacher)
	assert.Equal(t, newEmail, resp.Teacher.Email)
	assert.Equal(t, newUsername, resp.Teacher.Username)
	assert.Equal(t, pb.Gender_FEMALE, resp.Teacher.Gender)
}

// TestUpdateTeacher_NotFound tests updating non-existent teacher
func TestUpdateTeacher_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newEmail := "newemail@example.com"

	mock.ExpectExec("UPDATE Teacher SET").
		WithArgs(
			newEmail,
			fixtures.UpdatedBy,
			fixtures.TeacherID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Teacher WHERE id").
		WithArgs(fixtures.TeacherID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateTeacherRequest{
		Id:        fixtures.TeacherID1,
		Email:     &newEmail,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateTeacher_NoFieldsToUpdate tests update with no fields
func TestUpdateTeacher_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateTeacherRequest{
		Id:        fixtures.TeacherID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteTeacher_Success tests successful teacher deletion
func TestDeleteTeacher_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Teacher WHERE id").
		WithArgs(fixtures.TeacherID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteTeacherRequest{
		Id: fixtures.TeacherID1,
	}

	resp, err := h.DeleteTeacher(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteTeacher_NotFound tests deleting non-existent teacher
func TestDeleteTeacher_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Teacher WHERE id").
		WithArgs(fixtures.TeacherID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteTeacherRequest{
		Id: fixtures.TeacherID1,
	}

	resp, err := h.DeleteTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteTeacher_MissingID tests deletion with missing ID
func TestDeleteTeacher_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteTeacherRequest{
		Id: "",
	}

	resp, err := h.DeleteTeacher(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListTeachers_Success tests successful teachers listing
func TestListTeachers_Success(t *testing.T) {
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
		"id", "email", "username", "gender", "major_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.TeacherID1,
			fixtures.TeacherEmail1,
			fixtures.TeacherUsername1,
			"male",
			fixtures.TeacherMajorCode1,
			fixtures.TeacherSemesterCode1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.TeacherID2,
			fixtures.TeacherEmail2,
			fixtures.TeacherUsername2,
			"female",
			fixtures.TeacherMajorCode2,
			fixtures.TeacherSemesterCode2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Teacher").
		WillReturnRows(rows)

	req := &pb.ListTeachersRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListTeachers(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.Teachers, 2)
	assert.Equal(t, fixtures.TeacherEmail1, resp.Teachers[0].Email)
	assert.Equal(t, pb.Gender_MALE, resp.Teachers[0].Gender)
}

// TestListTeachers_Empty tests listing with no results
func TestListTeachers_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "email", "username", "gender", "major_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Teacher").
		WillReturnRows(rows)

	req := &pb.ListTeachersRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListTeachers(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.Teachers)
}
