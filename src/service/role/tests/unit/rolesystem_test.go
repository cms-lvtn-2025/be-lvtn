package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"thaily/src/service/role/handler"
	"thaily/src/service/role/tests/fixtures"

	pb "thaily/proto/role"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ==================== CREATE ROLESYSTEM TESTS ====================

func TestCreateRoleSystem_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateRoleSystemRequest()

	// Mock INSERT query - role enum is converted to string "teacher"
	mock.ExpectExec("INSERT INTO RoleSystem").
		WithArgs(
			sqlmock.AnyArg(), // id
			req.Title,
			req.TeacherCode,
			"teacher", // Role enum converted to string
			req.SemesterCode,
			req.Activate,
			req.CreatedBy,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SELECT query (GetRoleSystem is called after insert)
	// Column order: id, title, teacher_code, role, semester_code, activate, created_at, updated_at, created_by, updated_by
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem WHERE id = ?").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "teacher_code", "role", "semester_code", "activate",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			"test-id", req.Title, req.TeacherCode, "teacher", req.SemesterCode, req.Activate,
			time.Now(), time.Now(), req.CreatedBy, "",
		))

	// Execute
	resp, err := h.CreateRoleSystem(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.RoleSystem)
	assert.Equal(t, req.Title, resp.RoleSystem.Title)
	assert.Equal(t, pb.RoleType_TEACHER, resp.RoleSystem.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateRoleSystem_MissingTitle(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := &pb.CreateRoleSystemRequest{
		Title:        "",
		TeacherCode:  "T001",
		Role:         pb.RoleType_TEACHER,
		SemesterCode: "2024-1",
		Activate:     true,
		CreatedBy:    "test-user",
	}

	// Execute
	resp, err := h.CreateRoleSystem(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateRoleSystem_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestCreateRoleSystemRequest()

	// Mock INSERT with error
	mock.ExpectExec("INSERT INTO RoleSystem").
		WillReturnError(sql.ErrConnDone)

	// Execute
	resp, err := h.CreateRoleSystem(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GET ROLESYSTEM TESTS ====================

func TestGetRoleSystem_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	roleSystem := fixtures.GetTestRoleSystem()

	// Mock SELECT query - role is stored as string "teacher" in DB
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem WHERE id = ?").
		WithArgs(roleSystem.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "teacher_code", "role", "semester_code", "activate",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			roleSystem.Id, roleSystem.Title, roleSystem.TeacherCode, "teacher",
			roleSystem.SemesterCode, roleSystem.Activate,
			roleSystem.CreatedAt.AsTime(), roleSystem.UpdatedAt.AsTime(),
			roleSystem.CreatedBy, roleSystem.UpdatedBy,
		))

	// Execute
	resp, err := h.GetRoleSystem(context.Background(), &pb.GetRoleSystemRequest{Id: roleSystem.Id})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.RoleSystem)
	assert.Equal(t, roleSystem.Id, resp.RoleSystem.Id)
	assert.Equal(t, pb.RoleType_TEACHER, resp.RoleSystem.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoleSystem_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock SELECT with no rows
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem WHERE id = ?").
		WithArgs(fixtures.TestRoleSystemID1).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.GetRoleSystem(context.Background(), &pb.GetRoleSystemRequest{Id: fixtures.TestRoleSystemID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoleSystem_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.GetRoleSystem(context.Background(), &pb.GetRoleSystemRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== UPDATE ROLESYSTEM TESTS ====================

func TestUpdateRoleSystem_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateRoleSystemRequest()

	// Mock UPDATE - role enum is converted to "department_lecturer"
	mock.ExpectExec("UPDATE RoleSystem SET").
		WithArgs(
			*req.Title,
			*req.TeacherCode,
			"department_lecturer", // Role enum converted
			*req.SemesterCode,
			*req.Activate,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT for GetRoleSystem (called after update)
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem WHERE id = ?").
		WithArgs(req.Id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "teacher_code", "role", "semester_code", "activate",
			"created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			req.Id, *req.Title, *req.TeacherCode, "department_lecturer",
			*req.SemesterCode, *req.Activate,
			time.Now(), time.Now(), "test-user", req.UpdatedBy,
		))

	// Execute
	resp, err := h.UpdateRoleSystem(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.RoleSystem)
	assert.Equal(t, *req.Title, resp.RoleSystem.Title)
	assert.Equal(t, pb.RoleType_DEPARTMENT_LECTURER, resp.RoleSystem.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRoleSystem_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	req := fixtures.GetTestUpdateRoleSystemRequest()

	// Mock UPDATE
	mock.ExpectExec("UPDATE RoleSystem SET").
		WithArgs(
			*req.Title,
			*req.TeacherCode,
			"department_lecturer",
			*req.SemesterCode,
			*req.Activate,
			req.UpdatedBy,
			req.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT that returns NotFound
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem WHERE id = ?").
		WithArgs(req.Id).
		WillReturnError(sql.ErrNoRows)

	// Execute
	resp, err := h.UpdateRoleSystem(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRoleSystem_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	title := "Updated"
	req := &pb.UpdateRoleSystemRequest{
		Id:        "",
		Title:     &title,
		UpdatedBy: "test-user",
	}

	// Execute
	resp, err := h.UpdateRoleSystem(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== DELETE ROLESYSTEM TESTS ====================

func TestDeleteRoleSystem_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE
	mock.ExpectExec("DELETE FROM RoleSystem WHERE id = ?").
		WithArgs(fixtures.TestRoleSystemID1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	resp, err := h.DeleteRoleSystem(context.Background(), &pb.DeleteRoleSystemRequest{Id: fixtures.TestRoleSystemID1})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRoleSystem_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock DELETE with no rows
	mock.ExpectExec("DELETE FROM RoleSystem WHERE id = ?").
		WithArgs(fixtures.TestRoleSystemID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	resp, err := h.DeleteRoleSystem(context.Background(), &pb.DeleteRoleSystemRequest{Id: fixtures.TestRoleSystemID1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRoleSystem_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Execute
	resp, err := h.DeleteRoleSystem(context.Background(), &pb.DeleteRoleSystemRequest{Id: ""})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ==================== LIST ROLESYSTEMS TESTS ====================

func TestListRoleSystems_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)
	roleSystem1 := fixtures.GetTestRoleSystem()
	roleSystem2 := fixtures.GetTestRoleSystem2()

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM RoleSystem").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock SELECT - roles stored as strings in DB
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "teacher_code", "role", "semester_code", "activate",
			"created_at", "updated_at", "created_by", "updated_by",
		}).
			AddRow(
				roleSystem1.Id, roleSystem1.Title, roleSystem1.TeacherCode, "teacher",
				roleSystem1.SemesterCode, roleSystem1.Activate,
				roleSystem1.CreatedAt.AsTime(), roleSystem1.UpdatedAt.AsTime(),
				roleSystem1.CreatedBy, roleSystem1.UpdatedBy,
			).
			AddRow(
				roleSystem2.Id, roleSystem2.Title, roleSystem2.TeacherCode, "department_lecturer",
				roleSystem2.SemesterCode, roleSystem2.Activate,
				roleSystem2.CreatedAt.AsTime(), roleSystem2.UpdatedAt.AsTime(),
				roleSystem2.CreatedBy, roleSystem2.UpdatedBy,
			))

	// Execute
	req := fixtures.GetTestListRoleSystemsRequest()
	resp, err := h.ListRoleSystems(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.RoleSystems, 2)
	assert.Equal(t, int32(2), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListRoleSystems_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM RoleSystem").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock SELECT
	mock.ExpectQuery("SELECT (.+) FROM RoleSystem").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "teacher_code", "role", "semester_code", "activate",
			"created_at", "updated_at", "created_by", "updated_by",
		}))

	// Execute
	req := fixtures.GetTestListRoleSystemsRequest()
	resp, err := h.ListRoleSystems(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.RoleSystems, 0)
	assert.Equal(t, int32(0), resp.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListRoleSystems_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := handler.NewHandler(db)

	// Mock COUNT with error
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM RoleSystem").
		WillReturnError(sql.ErrConnDone)

	// Execute
	req := fixtures.GetTestListRoleSystemsRequest()
	resp, err := h.ListRoleSystems(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}
