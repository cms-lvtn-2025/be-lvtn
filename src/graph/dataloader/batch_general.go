package dataloader

import (
	"context"
	"thaily/proto/common"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// ============================================
// GENERAL ENTITY BATCH FUNCTIONS - BY ID
// ============================================

// createStudentByIdBatchFunc creates a batch function for loading students by ID
func createStudentByIdBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.Student] {
	return func(ctx context.Context, ids []string) (map[string]*model.Student, error) {
		result := make(map[string]*model.Student)

		for _, id := range ids {
			student, err := client.GetUserById(ctx, id)
			if err != nil {
				continue // Skip errors to avoid breaking the entire batch
			}
			result[id] = convert.PbStudentToModel(student.GetStudent())
		}

		return result, nil
	}
}

// createSemesterByIdBatchFunc creates a batch function for loading semesters by ID
func createSemesterByIdBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.Semester] {
	return func(ctx context.Context, ids []string) (map[string]*model.Semester, error) {
		result := make(map[string]*model.Semester)

		for _, id := range ids {
			semester, err := client.GetSemesterById(ctx, id)
			if err != nil {
				continue
			}
			result[id] = convert.PbSemesterToModel(semester.GetSemester())
		}

		return result, nil
	}
}

// createEnrollmentByIdBatchFunc creates a batch function for loading enrollments by ID
func createEnrollmentByIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Enrollment] {
	return func(ctx context.Context, ids []string) (map[string]*model.Enrollment, error) {
		result := make(map[string]*model.Enrollment)

		for _, id := range ids {
			enrollment, err := client.GetEnrollmentById(ctx, id)
			if err != nil {
				continue
			}
			result[id] = convert.PbEnrollmentToModel(enrollment.GetEnrollment())
		}

		return result, nil
	}
}

// createTopicCouncilByIdBatchFunc creates a batch function for loading topic councils by ID
func createTopicCouncilByIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.TopicCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.TopicCouncil, error) {
		result := make(map[string]*model.TopicCouncil)

		for _, id := range ids {
			topicCouncil, err := client.GetTopicCouncilById(ctx, id)
			if err != nil {
				continue
			}
			result[id] = convert.PbTopicCouncilToModel(topicCouncil.GetTopicCouncil())
		}

		return result, nil
	}
}

// createDefenceByIdBatchFunc creates a batch function for loading defences by ID
func createDefenceByIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.Defence] {
	return func(ctx context.Context, ids []string) (map[string]*model.Defence, error) {
		result := make(map[string]*model.Defence)

		for _, id := range ids {
			defence, err := client.GetDefenceById(ctx, id)
			if err != nil {
				continue
			}
			result[id] = convert.PbDefenceToModel(defence.GetDefence())
		}

		return result, nil
	}
}

// createCouncilByIdBatchFunc creates a batch function for loading councils by ID
func createCouncilByIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.Council] {
	return func(ctx context.Context, ids []string) (map[string]*model.Council, error) {
		result := make(map[string]*model.Council)

		for _, id := range ids {
			council, err := client.GetCouncilById(ctx, id)
			if err != nil {
				continue
			}
			result[id] = convert.PbCouncilToModel(council.GetCouncil())
		}

		return result, nil
	}
}

// ============================================
// ONE-TO-MANY RELATIONSHIP BATCH FUNCTIONS
// ============================================

// createEnrollmentsByStudentIdBatchFunc creates a batch function for loading enrollments by student ID
func createEnrollmentsByStudentIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Enrollment] {
	return func(ctx context.Context, studentIds []string) (map[string][]*model.Enrollment, error) {
		result := make(map[string][]*model.Enrollment)

		for _, studentId := range studentIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "student_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{studentId},
							},
						},
					},
				},
			}

			enrollments, err := client.GetEnrollmentBySearch(ctx, searchRequest)
			if err != nil {
				result[studentId] = []*model.Enrollment{}
				continue
			}

			result[studentId] = convert.PbEnrollmentsToModel(enrollments.GetEnrollments())
		}

		return result, nil
	}
}

// createRolesByTeacherIdBatchFunc creates a batch function for loading roles by teacher ID
func createRolesByTeacherIdBatchFunc(client *client.GRPCRole) BatchFunc[string, []*model.RoleSystem] {
	return func(ctx context.Context, teacherIds []string) (map[string][]*model.RoleSystem, error) {
		result := make(map[string][]*model.RoleSystem)

		for _, teacherId := range teacherIds {
			roles, err := client.GetAllRoleByTeacherId(ctx, teacherId)
			if err != nil {
				result[teacherId] = []*model.RoleSystem{}
				continue
			}

			result[teacherId] = convert.PbRoleSystemsToModel(roles.GetRoleSystems())
		}

		return result, nil
	}
}

// createMajorsByFacultyIdBatchFunc creates a batch function for loading majors by faculty ID
func createMajorsByFacultyIdBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, []*model.Major] {
	return func(ctx context.Context, facultyIds []string) (map[string][]*model.Major, error) {
		result := make(map[string][]*model.Major)

		for _, facultyId := range facultyIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "faculty_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{facultyId},
							},
						},
					},
				},
			}

			majors, err := client.GetMajorsBySearch(ctx, searchRequest)
			if err != nil {
				result[facultyId] = []*model.Major{}
				continue
			}

			result[facultyId] = convert.PbMajorsToModel(majors.GetMajors())
		}

		return result, nil
	}
}

// createTopicsByMajorIdBatchFunc creates a batch function for loading topics by major ID
func createTopicsByMajorIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Topic] {
	return func(ctx context.Context, majorIds []string) (map[string][]*model.Topic, error) {
		result := make(map[string][]*model.Topic)

		for _, majorId := range majorIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "major_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{majorId},
							},
						},
					},
				},
			}

			topics, err := client.GetTopicBySearch(ctx, searchRequest)
			if err != nil {
				result[majorId] = []*model.Topic{}
				continue
			}

			result[majorId] = convert.PbTopicsToModel(topics.GetTopics())
		}

		return result, nil
	}
}

// createStudentsBySemesterIdBatchFunc creates a batch function for loading students by semester ID
func createStudentsBySemesterIdBatchFunc(client *client.GRPCUser) BatchFunc[string, []*model.Student] {
	return func(ctx context.Context, semesterIds []string) (map[string][]*model.Student, error) {
		result := make(map[string][]*model.Student)

		for _, semesterId := range semesterIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "semester_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{semesterId},
							},
						},
					},
				},
			}

			students, err := client.GetStudentsBySearch(ctx, searchRequest)
			if err != nil {
				result[semesterId] = []*model.Student{}
				continue
			}

			result[semesterId] = convert.PbStudentsToModel(students.GetStudents())
		}

		return result, nil
	}
}

// createTeachersBySemesterIdBatchFunc creates a batch function for loading teachers by semester ID
func createTeachersBySemesterIdBatchFunc(client *client.GRPCUser) BatchFunc[string, []*model.Teacher] {
	return func(ctx context.Context, semesterIds []string) (map[string][]*model.Teacher, error) {
		result := make(map[string][]*model.Teacher)

		for _, semesterId := range semesterIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "semester_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{semesterId},
							},
						},
					},
				},
			}

			teachers, err := client.GetTeachersBySearch(ctx, searchRequest)
			if err != nil {
				result[semesterId] = []*model.Teacher{}
				continue
			}

			result[semesterId] = convert.PbTeachersToModel(teachers.GetTeachers())
		}

		return result, nil
	}
}

// createTopicsBySemesterIdBatchFunc creates a batch function for loading topics by semester ID
func createTopicsBySemesterIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Topic] {
	return func(ctx context.Context, semesterIds []string) (map[string][]*model.Topic, error) {
		result := make(map[string][]*model.Topic)

		for _, semesterId := range semesterIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "semester_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{semesterId},
							},
						},
					},
				},
			}

			topics, err := client.GetTopicBySearch(ctx, searchRequest)
			if err != nil {
				result[semesterId] = []*model.Topic{}
				continue
			}

			result[semesterId] = convert.PbTopicsToModel(topics.GetTopics())
		}

		return result, nil
	}
}

// createFilesByTopicIdBatchFunc creates a batch function for loading files by topic ID
func createFilesByTopicIdBatchFunc(client *client.GRPCfile) BatchFunc[string, []*model.File] {
	return func(ctx context.Context, topicIds []string) (map[string][]*model.File, error) {
		result := make(map[string][]*model.File)

		for _, topicId := range topicIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "topic_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{topicId},
							},
						},
					},
				},
			}

			files, err := client.GetFileBySearch(ctx, searchRequest)
			if err != nil {
				result[topicId] = []*model.File{}
				continue
			}

			result[topicId] = convert.PbFilesToModel(files.GetFiles())
		}

		return result, nil
	}
}

// createTopicCouncilsByTopicIdBatchFunc creates a batch function for loading topic councils by topic ID
func createTopicCouncilsByTopicIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.TopicCouncil] {
	return func(ctx context.Context, topicIds []string) (map[string][]*model.TopicCouncil, error) {
		result := make(map[string][]*model.TopicCouncil)

		for _, topicId := range topicIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "topic_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{topicId},
							},
						},
					},
				},
			}

			topicCouncils, err := client.GetTopicCouncilBySearch(ctx, searchRequest)
			if err != nil {
				result[topicId] = []*model.TopicCouncil{}
				continue
			}

			result[topicId] = convert.PbTopicCouncilsToModel(topicCouncils.GetTopicCouncils())
		}

		return result, nil
	}
}

// createEnrollmentsByTopicCouncilIdBatchFunc creates a batch function for loading enrollments by topic council ID
func createEnrollmentsByTopicCouncilIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Enrollment] {
	return func(ctx context.Context, topicCouncilIds []string) (map[string][]*model.Enrollment, error) {
		result := make(map[string][]*model.Enrollment)

		for _, topicCouncilId := range topicCouncilIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "topic_council_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{topicCouncilId},
							},
						},
					},
				},
			}

			enrollments, err := client.GetEnrollmentBySearch(ctx, searchRequest)
			if err != nil {
				result[topicCouncilId] = []*model.Enrollment{}
				continue
			}

			result[topicCouncilId] = convert.PbEnrollmentsToModel(enrollments.GetEnrollments())
		}

		return result, nil
	}
}

// createSupervisorsByTopicCouncilIdBatchFunc creates a batch function for loading supervisors by topic council ID
func createSupervisorsByTopicCouncilIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.TopicCouncilSupervisor] {
	return func(ctx context.Context, topicCouncilIds []string) (map[string][]*model.TopicCouncilSupervisor, error) {
		result := make(map[string][]*model.TopicCouncilSupervisor)

		for _, topicCouncilId := range topicCouncilIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "topic_council_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{topicCouncilId},
							},
						},
					},
				},
			}

			supervisors, err := client.GetTopicCouncilSupervisorBySearch(ctx, searchRequest)
			if err != nil {
				result[topicCouncilId] = []*model.TopicCouncilSupervisor{}
				continue
			}

			result[topicCouncilId] = convert.PbTopicCouncilSupervisorsToModel(supervisors.GetTopicCouncilSupervisors())
		}

		return result, nil
	}
}

// createDefencesByCouncilIdBatchFunc creates a batch function for loading defences by council ID
func createDefencesByCouncilIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.Defence] {
	return func(ctx context.Context, councilIds []string) (map[string][]*model.Defence, error) {
		result := make(map[string][]*model.Defence)

		for _, councilId := range councilIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "council_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{councilId},
							},
						},
					},
				},
			}

			defences, err := client.GetDefencesBySearch(ctx, searchRequest)
			if err != nil {
				result[councilId] = []*model.Defence{}
				continue
			}

			result[councilId] = convert.PbDefencesToModel(defences.GetDefences())
		}

		return result, nil
	}
}

// createTopicCouncilsByCouncilIdBatchFunc creates a batch function for loading topic councils by council ID
func createTopicCouncilsByCouncilIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.TopicCouncil] {
	return func(ctx context.Context, councilIds []string) (map[string][]*model.TopicCouncil, error) {
		result := make(map[string][]*model.TopicCouncil)

		for _, councilId := range councilIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "council_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{councilId},
							},
						},
					},
				},
			}

			topicCouncils, err := client.GetTopicCouncilBySearch(ctx, searchRequest)
			if err != nil {
				result[councilId] = []*model.TopicCouncil{}
				continue
			}

			result[councilId] = convert.PbTopicCouncilsToModel(topicCouncils.GetTopicCouncils())
		}

		return result, nil
	}
}

// createGradeDefencesByEnrollmentIdBatchFunc creates a batch function for loading grade defences by enrollment ID
func createGradeDefencesByEnrollmentIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefence] {
	return func(ctx context.Context, enrollmentIds []string) (map[string][]*model.GradeDefence, error) {
		result := make(map[string][]*model.GradeDefence)

		for _, enrollmentId := range enrollmentIds {
			searchRequest := &common.SearchRequest{
				Pagination: &common.Pagination{
					Page:     1,
					PageSize: 100,
				},
				Filters: []*common.FilterCriteria{
					{
						Criteria: &common.FilterCriteria_Condition{
							Condition: &common.FilterCondition{
								Field:    "enrollment_code",
								Operator: common.FilterOperator_EQUAL,
								Values:   []string{enrollmentId},
							},
						},
					},
				},
			}

			gradeDefences, err := client.GetGradeDefenceBySearch(ctx, searchRequest)
			if err != nil {
				result[enrollmentId] = []*model.GradeDefence{}
				continue
			}

			result[enrollmentId] = convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences())
		}

		return result, nil
	}
}
