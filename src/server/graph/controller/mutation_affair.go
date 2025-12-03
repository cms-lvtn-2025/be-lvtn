package controller

import (
	"context"
	"fmt"
	"time"

	pbAcademic "thaily/proto/academic"
	pbCouncil "thaily/proto/council"
	pbRole "thaily/proto/role"
	pbThesis "thaily/proto/thesis"
	pbUser "thaily/proto/user"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/dataloader"
	"thaily/src/server/graph/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================
// AFFAIR MUTATION CONTROLLER
// Handles: AffairMutation resolvers (Academic Affairs Staff)
// Full CRUD access to all entities
// ============================================

// ============================================
// USER MANAGEMENT MUTATIONS
// ============================================

// CreateTeacher creates a new teacher
func (c *Controller) CreateTeacher(ctx context.Context, input model.CreateTeacherInput) (*model.Teacher, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	// Convert gender
	var gender pbUser.Gender
	switch input.Gender {
	case model.GenderMale:
		gender = pbUser.Gender_MALE
	case model.GenderFemale:
		gender = pbUser.Gender_FEMALE
	default:
		gender = pbUser.Gender_MALE
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}

	// Create teacher via gRPC
	resp, err := c.user.CreateTeacher(ctx, &pbUser.CreateTeacherRequest{
		Id:           input.ID,
		Email:        input.Email,
		Username:     input.Username,
		Gender:       gender,
		MajorCode:    input.MajorCode,
		SemesterCode: input.SemesterCode,
		Msgv:         input.Msgv,
		CreatedBy:    createdBy,
	})
	if err != nil {
		return nil, err
	}

	// Create roles for the teacher if provided
	if len(input.Roles) > 0 {
		pbRoles := make([]pbRole.RoleType, 0, len(input.Roles))
		for _, role := range input.Roles {
			if role != nil {
				pbRoles = append(pbRoles, modelRoleToPbRoleType(*role))
			}
		}

		roleReq := &pbRole.CreateRoleSystemRequest{
			Title:        input.Username,
			TeacherCode:  input.Msgv,
			SemesterCode: input.SemesterCode,
			Activate:     true,
			CreatedBy:    createdBy,
		}

		_, err = c.role.CreateRoles(ctx, roleReq, pbRoles)
		if err != nil {
			// Log warning but don't fail teacher creation
			fmt.Printf("Warning: Failed to create roles for teacher %s: %v\n", input.Msgv, err)
		}
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateTeacherBySemester(input.SemesterCode)
	}

	return convert.PbTeacherToModel(resp.GetTeacher()), nil
}

// UpdateTeacher updates a teacher
func (c *Controller) UpdateTeacher(ctx context.Context, id string, input model.UpdateTeacherInput) (*model.Teacher, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbUser.UpdateTeacherRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Email != nil {
		req.Email = input.Email
	}
	if input.Username != nil {
		req.Username = input.Username
	}
	if input.Gender != nil {
		var gender pbUser.Gender
		switch *input.Gender {
		case model.GenderMale:
			gender = pbUser.Gender_MALE
		case model.GenderFemale:
			gender = pbUser.Gender_FEMALE
		}
		req.Gender = &gender
	}
	if input.MajorCode != nil {
		req.MajorCode = input.MajorCode
	}
	if input.SemesterCode != nil {
		req.SemesterCode = input.SemesterCode
	}
	if input.Msgv != nil {
		req.Msgv = input.Msgv
	}

	resp, err := c.user.UpdateTeacher(ctx, req)
	if err != nil {
		return nil, err
	}

	// Handle roles update if provided
	if input.Roles != nil {
		teacher := resp.GetTeacher()
		if teacher == nil {
			return nil, fmt.Errorf("teacher not found after update")
		}

		teacherCode := teacher.Id
		semesterCode := teacher.SemesterCode
		if input.SemesterCode != nil {
			semesterCode = *input.SemesterCode
		}

		// Get existing roles
		existingRolesResp, err := c.role.GetAllRoleByTeacherId(ctx, teacherCode)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing roles: %v", err)
		}

		// Build input role set
		inputRoleSet := make(map[pbRole.RoleType]bool)
		for _, role := range input.Roles {
			if role != nil {
				inputRoleSet[modelRoleToPbRoleType(*role)] = true
			}
		}

		// Build existing role map (roleType -> roleId)
		existingRoleMap := make(map[pbRole.RoleType]string)
		if existingRolesResp != nil && existingRolesResp.RoleSystems != nil {
			for _, role := range existingRolesResp.RoleSystems {
				if role != nil {
					existingRoleMap[role.Role] = role.Id
				}
			}
		}

		// Find roles to create
		rolesToCreate := make([]pbRole.RoleType, 0)
		for roleType := range inputRoleSet {
			if _, exists := existingRoleMap[roleType]; !exists {
				rolesToCreate = append(rolesToCreate, roleType)
			}
		}

		// Find roles to delete
		rolesToDelete := make([]string, 0)
		for roleType, roleId := range existingRoleMap {
			if !inputRoleSet[roleType] {
				rolesToDelete = append(rolesToDelete, roleId)
			}
		}

		// Create new roles
		if len(rolesToCreate) > 0 {
			roleReq := &pbRole.CreateRoleSystemRequest{
				Title:        teacher.Username,
				TeacherCode:  teacherCode,
				SemesterCode: semesterCode,
				Activate:     true,
				CreatedBy:    updatedBy,
			}

			_, err = c.role.CreateRoles(ctx, roleReq, rolesToCreate)
			if err != nil {
				return nil, fmt.Errorf("failed to create roles: %v", err)
			}
		}

		// Delete old roles
		for _, roleId := range rolesToDelete {
			_, err = c.role.DeleteRole(ctx, roleId)
			if err != nil {
				fmt.Printf("Warning: Failed to delete role %s: %v\n", roleId, err)
			}
		}
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateTeacher(id)
		if resp.GetTeacher() != nil {
			loaders.InvalidateTeacherBySemester(resp.GetTeacher().GetSemesterCode())
		}
	}

	return convert.PbTeacherToModel(resp.GetTeacher()), nil
}

// DeleteTeacher deletes a teacher
func (c *Controller) DeleteTeacher(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	// Get teacher info before delete for cache invalidation
	teacher, _ := c.user.GetTeacherById(ctx, id)

	_, err = c.user.DeleteTeacher(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateTeacher(id)
		if teacher != nil && teacher.GetTeacher() != nil {
			loaders.InvalidateTeacherBySemester(teacher.GetTeacher().GetSemesterCode())
		}
	}

	return true, nil
}

// CreateStudent creates a new student
func (c *Controller) CreateStudent(ctx context.Context, input model.CreateStudentInput) (*model.Student, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}

	// Convert gender
	var gender pbUser.Gender
	switch input.Gender {
	case model.GenderMale:
		gender = pbUser.Gender_MALE
	case model.GenderFemale:
		gender = pbUser.Gender_FEMALE
	default:
		gender = pbUser.Gender_MALE
	}

	req := &pbUser.CreateStudentRequest{
		Id:           input.ID,
		Email:        input.Email,
		Username:     input.Username,
		Gender:       &gender,
		MajorCode:    input.MajorCode,
		SemesterCode: input.SemesterCode,
		Mssv:         input.Mssv,
		CreatedBy:    createdBy,
	}

	if input.Phone != "" {
		req.Phone = &input.Phone
	}
	if input.ClassCode != nil && *input.ClassCode != "" {
		req.ClassCode = *input.ClassCode
	}

	resp, err := c.user.CreateStudent(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateStudentBySemester(input.SemesterCode)
	}

	return convert.PbStudentToModel(resp.GetStudent()), nil
}

// UpdateStudent updates a student
func (c *Controller) UpdateStudent(ctx context.Context, id string, input model.UpdateStudentInput) (*model.Student, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbUser.UpdateStudentRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Email != nil {
		req.Email = input.Email
	}
	if input.Phone != nil {
		req.Phone = input.Phone
	}
	if input.Username != nil {
		req.Username = input.Username
	}
	if input.Gender != nil {
		var gender pbUser.Gender
		switch *input.Gender {
		case model.GenderMale:
			gender = pbUser.Gender_MALE
		case model.GenderFemale:
			gender = pbUser.Gender_FEMALE
		}
		req.Gender = &gender
	}
	if input.MajorCode != nil {
		req.MajorCode = input.MajorCode
	}
	if input.ClassCode != nil {
		req.ClassCode = input.ClassCode
	}
	if input.SemesterCode != nil {
		req.SemesterCode = input.SemesterCode
	}
	if input.Mssv != nil {
		req.Mssv = input.Mssv
	}

	resp, err := c.user.UpdateStudent(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateStudent(id)
		if resp.GetStudent() != nil {
			loaders.InvalidateStudentBySemester(resp.GetStudent().GetSemesterCode())
		}
	}

	return convert.PbStudentToModel(resp.GetStudent()), nil
}

// DeleteStudent deletes a student
func (c *Controller) DeleteStudent(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	// Get student info before delete for cache invalidation
	student, _ := c.user.GetUserById(ctx, id)

	_, err = c.user.DeleteStudent(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateStudent(id)
		if student != nil && student.GetStudent() != nil {
			loaders.InvalidateStudentBySemester(student.GetStudent().GetSemesterCode())
		}
	}

	return true, nil
}

// ============================================
// ACADEMIC MANAGEMENT MUTATIONS
// ============================================

// CreateSemester creates a new semester
func (c *Controller) CreateSemester(ctx context.Context, input model.CreateSemesterInput) (*model.Semester, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}

	resp, err := c.academic.CreateSemester(ctx, &pbAcademic.CreateSemesterRequest{
		Title:     input.Title,
		CreatedBy: createdBy,
	})
	if err != nil {
		return nil, err
	}

	// No cache invalidation needed for create - new semester won't be in cache

	return convert.PbSemesterToModel(resp.GetSemester()), nil
}

// UpdateSemester updates a semester
func (c *Controller) UpdateSemester(ctx context.Context, id string, input model.UpdateSemesterInput) (*model.Semester, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbAcademic.UpdateSemesterRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Title != nil {
		req.Title = input.Title
	}

	resp, err := c.academic.UpdateSemester(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateSemester(id)
	}

	return convert.PbSemesterToModel(resp.GetSemester()), nil
}

// DeleteSemester deletes a semester
func (c *Controller) DeleteSemester(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	_, err = c.academic.DeleteSemester(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateSemester(id)
	}

	return true, nil
}

// CreateMajor creates a new major
func (c *Controller) CreateMajor(ctx context.Context, input model.CreateMajorInput) (*model.Major, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}

	resp, err := c.academic.CreateMajor(ctx, &pbAcademic.CreateMajorRequest{
		Id:          input.ID,
		Title:       input.Title,
		FacultyCode: input.FacultyCode,
		Ms:          input.Ms,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, err
	}

	// Invalidate cache - new major affects faculty's major list
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateMajorByFaculty(input.FacultyCode)
	}

	return convert.PbMajorToModel(resp.GetMajor()), nil
}

// UpdateMajor updates a major
func (c *Controller) UpdateMajor(ctx context.Context, id string, input model.UpdateMajorInput) (*model.Major, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbAcademic.UpdateMajorRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Title != nil {
		req.Title = input.Title
	}
	if input.FacultyCode != nil {
		req.FacultyCode = input.FacultyCode
	}
	if input.Ms != nil {
		req.Ms = input.Ms
	}

	resp, err := c.academic.UpdateMajor(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateMajor(id)
		if resp.GetMajor() != nil {
			loaders.InvalidateMajorByFaculty(resp.GetMajor().GetFacultyCode())
		}
	}

	return convert.PbMajorToModel(resp.GetMajor()), nil
}

// DeleteMajor deletes a major
func (c *Controller) DeleteMajor(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	// Get major info before delete for cache invalidation
	major, _ := c.academic.GetMajorById(ctx, id)

	_, err = c.academic.DeleteMajor(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateMajor(id)
		if major != nil && major.GetMajor() != nil {
			loaders.InvalidateMajorByFaculty(major.GetMajor().GetFacultyCode())
		}
	}

	return true, nil
}

// CreateFaculty creates a new faculty
func (c *Controller) CreateFaculty(ctx context.Context, input model.CreateFacultyInput) (*model.Faculty, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}

	resp, err := c.academic.CreateFaculty(ctx, &pbAcademic.CreateFacultyRequest{
		Id:        input.ID,
		Title:     input.Title,
		Ms:        input.Ms,
		CreatedBy: createdBy,
	})
	if err != nil {
		return nil, err
	}

	return convert.PbFacultyToModel(resp.GetFaculty()), nil
}

// UpdateFaculty updates a faculty
func (c *Controller) UpdateFaculty(ctx context.Context, id string, input model.UpdateFacultyInput) (*model.Faculty, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbAcademic.UpdateFacultyRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Title != nil {
		req.Title = input.Title
	}
	if input.Ms != nil {
		req.Ms = input.Ms
	}

	resp, err := c.academic.UpdateFaculty(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFacultyToModel(resp.GetFaculty()), nil
}

// DeleteFaculty deletes a faculty
func (c *Controller) DeleteFaculty(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	_, err = c.academic.DeleteFaculty(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ============================================
// COUNCIL MANAGEMENT MUTATIONS
// ============================================

// ApproveCouncil approves a council by setting time start
func (c *Controller) ApproveCouncil(ctx context.Context, id string, timeStart time.Time) (*model.Council, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	ts := timestamppb.New(timeStart)
	req := &pbCouncil.UpdateCouncilRequest{
		Id:        id,
		TimeStart: ts,
		UpdatedBy: updatedBy,
	}

	resp, err := c.council.UpdateCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateCouncil(id)
	}

	return convert.PbCouncilToModel(resp.GetCouncil()), nil
}

// UpdateCouncil updates a council
func (c *Controller) UpdateCouncil(ctx context.Context, id string, input model.UpdateCouncilInput) (*model.Council, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbCouncil.UpdateCouncilRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Title != nil {
		req.Title = input.Title
	}
	if input.TimeStart != nil {
		ts := timestamppb.New(*input.TimeStart)
		req.TimeStart = ts
	}

	resp, err := c.council.UpdateCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateCouncil(id)
	}

	return convert.PbCouncilToModel(resp.GetCouncil()), nil
}

// DeleteCouncil deletes a council
func (c *Controller) DeleteCouncil(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	_, err = c.council.DeleteCouncil(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateCouncil(id)
	}

	return true, nil
}

// ============================================
// TOPIC MANAGEMENT MUTATIONS
// ============================================

// ApproveTopic approves topic stage 2 (final approval)
func (c *Controller) ApproveTopic(ctx context.Context, id string) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	status := pbThesis.TopicStatus_APPROVED_2
	req := &pbThesis.UpdateTopicRequest{
		Id:        id,
		Status:    &status,
		UpdatedBy: updatedBy,
	}

	resp, err := c.thesis.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetTopic() != nil {
		topic := resp.GetTopic()
		loaders.InvalidateTopic(id, topic.GetMajorCode(), topic.GetSemesterCode())
	}

	return convert.PbTopicToModel(resp.GetTopic()), nil
}

// RejectTopic rejects a topic
func (c *Controller) RejectTopic(ctx context.Context, id string, reason *string) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	status := pbThesis.TopicStatus_REJECTED
	req := &pbThesis.UpdateTopicRequest{
		Id:        id,
		Status:    &status,
		UpdatedBy: updatedBy,
	}

	resp, err := c.thesis.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetTopic() != nil {
		topic := resp.GetTopic()
		loaders.InvalidateTopic(id, topic.GetMajorCode(), topic.GetSemesterCode())
	}

	return convert.PbTopicToModel(resp.GetTopic()), nil
}

// UpdateTopic updates a topic
func (c *Controller) UpdateTopic(ctx context.Context, id string, input model.UpdateTopicInput) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	req := &pbThesis.UpdateTopicRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Title != nil {
		req.Title = input.Title
	}
	if input.Status != nil {
		var status pbThesis.TopicStatus
		switch *input.Status {
		case model.TopicStatusApproved1:
			status = pbThesis.TopicStatus_APPROVED_1
		case model.TopicStatusApproved2:
			status = pbThesis.TopicStatus_APPROVED_2
		case model.TopicStatusRejected:
			status = pbThesis.TopicStatus_REJECTED
		}
		req.Status = &status
	}
	if input.PercentStage1 != nil {
		percentStage1 := int32(*input.PercentStage1)
		req.PercentStage_1 = &percentStage1
	}
	if input.PercentStage2 != nil {
		percentStage2 := int32(*input.PercentStage2)
		req.PercentStage_2 = &percentStage2
	}

	resp, err := c.thesis.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetTopic() != nil {
		topic := resp.GetTopic()
		loaders.InvalidateTopic(id, topic.GetMajorCode(), topic.GetSemesterCode())
	}

	return convert.PbTopicToModel(resp.GetTopic()), nil
}

// DeleteTopic deletes a topic
func (c *Controller) DeleteTopic(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires academic affairs staff role")
	}

	// Get topic info before delete for cache invalidation
	topic, _ := c.thesis.GetTopicById(ctx, id)

	_, err = c.thesis.DeleteTopic(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		majorCode := ""
		semesterCode := ""
		if topic != nil && topic.GetTopic() != nil {
			majorCode = topic.GetTopic().GetMajorCode()
			semesterCode = topic.GetTopic().GetSemesterCode()
		}
		loaders.InvalidateTopic(id, majorCode, semesterCode)
	}

	return true, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// modelRoleToPbRoleType converts model.RoleSystemRole to pb.RoleType
func modelRoleToPbRoleType(role model.RoleSystemRole) pbRole.RoleType {
	switch role {
	case model.RoleSystemRoleAcademicAffairsStaff:
		return pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF
	case model.RoleSystemRoleDepartmentLecturer:
		return pbRole.RoleType_DEPARTMENT_LECTURER
	case model.RoleSystemRoleTeacher:
		return pbRole.RoleType_TEACHER
	default:
		return pbRole.RoleType_TEACHER
	}
}
