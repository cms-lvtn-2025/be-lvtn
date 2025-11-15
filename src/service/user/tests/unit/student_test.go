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

// TestCreateStudent_Success tests successful student creation
func TestCreateStudent_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	gender := pb.Gender_MALE

	// Mock INSERT query
	mock.ExpectExec("INSERT INTO Student").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			fixtures.StudentEmail1,
			fixtures.StudentPhone1,
			fixtures.StudentUsername1,
			"male", // gender enum as string
			fixtures.StudentMajorCode1,
			fixtures.StudentClassCode1,
			fixtures.StudentSemesterCode1,
			fixtures.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetStudent
	rows := sqlmock.NewRows([]string{
		"id", "email", "phone", "username", "gender", "major_code", "class_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.StudentID1,
		fixtures.StudentEmail1,
		fixtures.StudentPhone1,
		fixtures.StudentUsername1,
		"male",
		fixtures.StudentMajorCode1,
		fixtures.StudentClassCode1,
		fixtures.StudentSemesterCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Student WHERE id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := &pb.CreateStudentRequest{
		Email:        fixtures.StudentEmail1,
		Phone:        &fixtures.StudentPhone1,
		Username:     fixtures.StudentUsername1,
		Gender:       &gender,
		MajorCode:    fixtures.StudentMajorCode1,
		ClassCode:    fixtures.StudentClassCode1,
		SemesterCode: fixtures.StudentSemesterCode1,
		CreatedBy:    fixtures.CreatedBy,
	}

	resp, err := h.CreateStudent(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Student)
	assert.Equal(t, fixtures.StudentEmail1, resp.Student.Email)
}

// TestCreateStudent_MissingEmail tests creation with missing email
func TestCreateStudent_MissingEmail(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateStudentRequest{
		Email:        "",
		Username:     fixtures.StudentUsername1,
		MajorCode:    fixtures.StudentMajorCode1,
		ClassCode:    fixtures.StudentClassCode1,
		SemesterCode: fixtures.StudentSemesterCode1,
		CreatedBy:    fixtures.CreatedBy,
	}

	resp, err := h.CreateStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestCreateStudent_MissingUsername tests creation with missing username
func TestCreateStudent_MissingUsername(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.CreateStudentRequest{
		Email:        fixtures.StudentEmail1,
		Username:     "",
		MajorCode:    fixtures.StudentMajorCode1,
		ClassCode:    fixtures.StudentClassCode1,
		SemesterCode: fixtures.StudentSemesterCode1,
		CreatedBy:    fixtures.CreatedBy,
	}

	resp, err := h.CreateStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestGetStudent_Success tests successful student retrieval
func TestGetStudent_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "email", "phone", "username", "gender", "major_code", "class_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.StudentID1,
		fixtures.StudentEmail1,
		fixtures.StudentPhone1,
		fixtures.StudentUsername1,
		"male",
		fixtures.StudentMajorCode1,
		fixtures.StudentClassCode1,
		fixtures.StudentSemesterCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)

	mock.ExpectQuery("SELECT (.+) FROM Student WHERE id").
		WithArgs(fixtures.StudentID1).
		WillReturnRows(rows)

	req := &pb.GetStudentRequest{
		Id: fixtures.StudentID1,
	}

	resp, err := h.GetStudent(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Student)
	assert.Equal(t, fixtures.StudentID1, resp.Student.Id)
	assert.Equal(t, fixtures.StudentEmail1, resp.Student.Email)
	assert.Equal(t, pb.Gender_MALE, resp.Student.Gender)
}

// TestGetStudent_NotFound tests student not found
func TestGetStudent_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM Student WHERE id").
		WithArgs(fixtures.StudentID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.GetStudentRequest{
		Id: fixtures.StudentID1,
	}

	resp, err := h.GetStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetStudent_MissingID tests getting student with missing ID
func TestGetStudent_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.GetStudentRequest{
		Id: "",
	}

	resp, err := h.GetStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpdateStudent_Success tests successful student update
func TestUpdateStudent_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newEmail := "newemail@example.com"
	newUsername := "new_username"
	newGender := pb.Gender_FEMALE

	// Mock UPDATE query - only email, username, gender are provided
	mock.ExpectExec("UPDATE Student SET").
		WithArgs(
			newEmail,
			newUsername,
			"female", // gender enum as string
			fixtures.UpdatedBy,
			fixtures.StudentID1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT for GetStudent
	rows := sqlmock.NewRows([]string{
		"id", "email", "phone", "username", "gender", "major_code", "class_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).AddRow(
		fixtures.StudentID1,
		newEmail,
		fixtures.StudentPhone1,
		newUsername,
		"female",
		fixtures.StudentMajorCode1,
		fixtures.StudentClassCode1,
		fixtures.StudentSemesterCode1,
		time.Now(),
		time.Now(),
		fixtures.CreatedBy,
		fixtures.UpdatedBy,
	)
	mock.ExpectQuery("SELECT (.+) FROM Student WHERE id").
		WithArgs(fixtures.StudentID1).
		WillReturnRows(rows)

	req := &pb.UpdateStudentRequest{
		Id:        fixtures.StudentID1,
		Email:     &newEmail,
		Username:  &newUsername,
		Gender:    &newGender,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateStudent(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Student)
	assert.Equal(t, newEmail, resp.Student.Email)
	assert.Equal(t, newUsername, resp.Student.Username)
	assert.Equal(t, pb.Gender_FEMALE, resp.Student.Gender)
}

// TestUpdateStudent_NotFound tests updating non-existent student
func TestUpdateStudent_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	newEmail := "newemail@example.com"

	mock.ExpectExec("UPDATE Student SET").
		WithArgs(
			newEmail,
			fixtures.UpdatedBy,
			fixtures.StudentID1,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT (.+) FROM Student WHERE id").
		WithArgs(fixtures.StudentID1).
		WillReturnError(sql.ErrNoRows)

	req := &pb.UpdateStudentRequest{
		Id:        fixtures.StudentID1,
		Email:     &newEmail,
		UpdatedBy: fixtures.UpdatedBy,
	}

	resp, err := h.UpdateStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// TestUpdateStudent_NoFieldsToUpdate tests update with no fields
func TestUpdateStudent_NoFieldsToUpdate(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.UpdateStudentRequest{
		Id:        fixtures.StudentID1,
		UpdatedBy: fixtures.UpdatedBy,
		// No fields to update
	}

	resp, err := h.UpdateStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestDeleteStudent_Success tests successful student deletion
func TestDeleteStudent_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Student WHERE id").
		WithArgs(fixtures.StudentID1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.DeleteStudentRequest{
		Id: fixtures.StudentID1,
	}

	resp, err := h.DeleteStudent(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

// TestDeleteStudent_NotFound tests deleting non-existent student
func TestDeleteStudent_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Student WHERE id").
		WithArgs(fixtures.StudentID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := &pb.DeleteStudentRequest{
		Id: fixtures.StudentID1,
	}

	resp, err := h.DeleteStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestDeleteStudent_MissingID tests deletion with missing ID
func TestDeleteStudent_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	req := &pb.DeleteStudentRequest{
		Id: "",
	}

	resp, err := h.DeleteStudent(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListStudents_Success tests successful students listing
func TestListStudents_Success(t *testing.T) {
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
		"id", "email", "phone", "username", "gender", "major_code", "class_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	}).
		AddRow(
			fixtures.StudentID1,
			fixtures.StudentEmail1,
			fixtures.StudentPhone1,
			fixtures.StudentUsername1,
			"male",
			fixtures.StudentMajorCode1,
			fixtures.StudentClassCode1,
			fixtures.StudentSemesterCode1,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		).
		AddRow(
			fixtures.StudentID2,
			fixtures.StudentEmail2,
			fixtures.StudentPhone2,
			fixtures.StudentUsername2,
			"female",
			fixtures.StudentMajorCode2,
			fixtures.StudentClassCode2,
			fixtures.StudentSemesterCode2,
			time.Now(),
			time.Now(),
			fixtures.CreatedBy,
			fixtures.UpdatedBy,
		)

	mock.ExpectQuery("SELECT (.+) FROM Student").
		WillReturnRows(rows)

	req := &pb.ListStudentsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListStudents(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.Students, 2)
	assert.Equal(t, fixtures.StudentEmail1, resp.Students[0].Email)
	assert.Equal(t, pb.Gender_MALE, resp.Students[0].Gender)
}

// TestListStudents_Empty tests listing with no results
func TestListStudents_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{
		"id", "email", "phone", "username", "gender", "major_code", "class_code", "semester_code",
		"created_at", "updated_at", "created_by", "updated_by",
	})
	mock.ExpectQuery("SELECT (.+) FROM Student").
		WillReturnRows(rows)

	req := &pb.ListStudentsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}

	resp, err := h.ListStudents(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.Students)
}
