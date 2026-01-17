package dataloader

import (
	"context"
	"fmt"
	"thaily/proto/common"
	"thaily/src/server/client"
	"thaily/src/server/graph/convert"
	convert2 "thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// GENERAL ENTITY BATCH FUNCTIONS - BY ID
// ============================================

// createStudentByIdBatchFunc creates a batch function for loading students by ID
func createStudentByIdBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.Student] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Student, error) {

		if len(ids) == 0 {
			return make(map[string]*model.Student), nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})
		page := int32(1)
		pageSize := int32(len(ids) * 10)
		result := make(map[string]*model.Student)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		students, _ := client.GetStudentsBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, student := range students.GetStudents() {
			result[student.GetId()] = convert.PbStudentToModel(student)
		}

		return result, nil
	}
}

// createSemesterByIdBatchFunc creates a batch function for loading semesters by ID
func createSemesterByIdBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.Semester] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Semester, error) {
		result := make(map[string]*model.Semester)

		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		semesters, _ := client.GetSemestersBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, semester := range semesters.GetSemesters() {
			result[semester.GetId()] = convert.PbSemesterToModel(semester)
		}

		return result, nil
	}
}

// createEnrollmentByIdBatchFunc creates a batch function for loading enrollments by ID
func createEnrollmentByIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Enrollment] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Enrollment, error) {
		result := make(map[string]*model.Enrollment)

		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		enrollments, _ := client.GetEnrollmentBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))
		for _, enrollment := range enrollments.GetEnrollments() {
			result[enrollment.GetId()] = convert2.PbEnrollmentToModel(enrollment)
		}
		return result, nil
	}
}

// createTopicCouncilByIdBatchFunc creates a batch function for loading topic councils by ID
func createTopicCouncilByIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.TopicCouncil] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.TopicCouncil, error) {
		result := make(map[string]*model.TopicCouncil)

		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})
		page := int32(1)
		pageSize := int32(len(ids) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		topicCouncils, _ := client.GetTopicCouncilBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))
		for _, topicCouncil := range topicCouncils.GetTopicCouncils() {
			result[topicCouncil.GetId()] = convert2.PbTopicCouncilToModel(topicCouncil)
		}

		return result, nil
	}
}

// createDefenceByIdBatchFunc creates a batch function for loading defences by ID
func createDefenceByIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.Defence] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Defence, error) {
		result := make(map[string]*model.Defence)

		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		defences, _ := client.GetDefencesBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))
		for _, defence := range defences.GetDefences() {
			result[defence.GetId()] = convert.PbDefenceToModel(defence)
		}

		return result, nil

	}
}

// createCouncilByIdBatchFunc creates a batch function for loading councils by ID
func createCouncilByIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.Council] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Council, error) {
		result := make(map[string]*model.Council)

		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})
		page := int32(1)
		pageSize := int32(len(ids) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}
		councils, _ := client.GetCouncilBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, council := range councils.GetCouncils() {
			result[council.GetId()] = convert.PbCouncilToModel(council)
		}

		return result, nil
	}
}

// ============================================
// ONE-TO-MANY RELATIONSHIP BATCH FUNCTIONS
// ============================================

// createEnrollmentsByStudentIdBatchFunc creates a batch function for loading enrollments by student ID
func createEnrollmentsByStudentIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Enrollment] {
	return func(ctx context.Context, studentIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Enrollment, error) {
		result := make(map[string][]*model.Enrollment)
		if len(studentIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "student_code",
				Operator: model.FilterOperatorIn,
				Values:   studentIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(studentIds) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		enrollments, _ := client.GetEnrollmentBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))

		for _, enrollment := range enrollments.GetEnrollments() {
			studentId := enrollment.GetStudentCode()
			result[studentId] = append(result[studentId], convert2.PbEnrollmentToModel(enrollment))
		}

		return result, nil
	}
}

// createRolesByTeacherIdBatchFunc creates a batch function for loading roles by teacher ID
func createRolesByTeacherIdBatchFunc(client *client.GRPCRole) BatchFunc[string, []*model.RoleSystem] {
	return func(ctx context.Context, teacherIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.RoleSystem, error) {
		result := make(map[string][]*model.RoleSystem)

		if len(teacherIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "teacher_code",
				Operator: model.FilterOperatorIn,
				Values:   teacherIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(teacherIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		roles, _ := client.GetRoleBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		fmt.Println(roles)

		for _, role := range roles.GetRoleSystems() {
			teacherId := role.GetTeacherCode()
			result[teacherId] = append(result[teacherId], convert2.PbRoleSystemToModel(role))
		}

		return result, nil
	}
}

// createMajorsByFacultyIdBatchFunc creates a batch function for loading majors by faculty ID
func createMajorsByFacultyIdBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, []*model.Major] {
	return func(ctx context.Context, facultyIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Major, error) {
		result := make(map[string][]*model.Major)

		if len(facultyIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "faculty_code",
				Operator: model.FilterOperatorIn,
				Values:   facultyIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(facultyIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		majors, _ := client.GetMajorsBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))
		for _, major := range majors.GetMajors() {
			facultyId := major.GetFacultyCode()
			result[facultyId] = append(result[facultyId], convert.PbMajorToModel(major))
		}

		return result, nil
	}
}

// createTopicsByMajorIdBatchFunc creates a batch function for loading topics by major ID
func createTopicsByMajorIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Topic] {
	return func(ctx context.Context, majorIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Topic, error) {
		result := make(map[string][]*model.Topic)

		if len(majorIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "major_code",
				Operator: model.FilterOperatorIn,
				Values:   majorIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(majorIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}
		topics, _ := client.GetTopicBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))
		for _, topic := range topics.GetTopics() {
			majorId := topic.GetMajorCode()
			result[majorId] = append(result[majorId], convert2.PbTopicToModel(topic))
		}

		return result, nil
	}
}

// createStudentsBySemesterIdBatchFunc creates a batch function for loading students by semester ID
func createStudentsBySemesterIdBatchFunc(client *client.GRPCUser) BatchFunc[string, []*model.Student] {
	return func(ctx context.Context, semesterIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Student, error) {
		result := make(map[string][]*model.Student)

		if len(semesterIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "semester_code",
				Operator: model.FilterOperatorIn,
				Values:   semesterIds,
			},
		})
		page := int32(1)
		pageSize := int32(len(semesterIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		students, _ := client.GetStudentsBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, student := range students.GetStudents() {
			semesterId := student.GetSemesterCode()
			result[semesterId] = append(result[semesterId], convert.PbStudentToModel(student))
		}
		return result, nil
	}
}

// createTeachersBySemesterIdBatchFunc creates a batch function for loading teachers by semester ID
func createTeachersBySemesterIdBatchFunc(client *client.GRPCUser) BatchFunc[string, []*model.Teacher] {
	return func(ctx context.Context, semesterIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Teacher, error) {
		result := make(map[string][]*model.Teacher)

		if len(semesterIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "semester_code",
				Operator: model.FilterOperatorIn,
				Values:   semesterIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(semesterIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		teachers, _ := client.GetTeachersBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))
		for _, teacher := range teachers.GetTeachers() {
			semesterId := teacher.GetSemesterCode()
			result[semesterId] = append(result[semesterId], convert.PbTeacherToModel(teacher))
		}
		return result, nil
	}
}

// createTopicsBySemesterIdBatchFunc creates a batch function for loading topics by semester ID
func createTopicsBySemesterIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Topic] {
	return func(ctx context.Context, semesterIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Topic, error) {
		result := make(map[string][]*model.Topic)

		if len(semesterIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "semester_code",
				Operator: model.FilterOperatorIn,
				Values:   semesterIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(semesterIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}
		topics, _ := client.GetTopicBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))

		for _, topic := range topics.GetTopics() {
			semesterId := topic.GetSemesterCode()
			result[semesterId] = append(result[semesterId], convert2.PbTopicToModel(topic))
		}

		return result, nil
	}
}

// createFilesByTopicIdBatchFunc creates a batch function for loading files by topic ID
func createFilesByTopicIdBatchFunc(client *client.GRPCfile) BatchFunc[string, []*model.File] {
	return func(ctx context.Context, topicIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.File, error) {
		result := make(map[string][]*model.File)

		if len(topicIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "table_id",
				Operator: model.FilterOperatorIn,
				Values:   topicIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(topicIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		files, _ := client.GetFileBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, file := range files.GetFiles() {
			topicId := file.GetTableId()
			result[topicId] = append(result[topicId], convert.PbFileToModel(file))
		}

		return result, nil
	}
}

// createFilesByTopicIdBatchFunc creates a batch function for loading files by topic ID
func createFilesByMidtermIdBatchFunc(client *client.GRPCfile) BatchFunc[string, []*model.File] {
	return func(ctx context.Context, midtermIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.File, error) {
		result := make(map[string][]*model.File)

		if len(midtermIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "table_id",
				Operator: model.FilterOperatorIn,
				Values:   midtermIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(midtermIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		files, _ := client.GetFileBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, file := range files.GetFiles() {
			midtermId := file.GetTableId()
			result[midtermId] = append(result[midtermId], convert.PbFileToModel(file))
		}

		return result, nil
	}
}

func createFilesByFinalIdBatchFunc(client *client.GRPCfile) BatchFunc[string, []*model.File] {
	return func(ctx context.Context, finalIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.File, error) {
		result := make(map[string][]*model.File)

		if len(finalIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "table_id",
				Operator: model.FilterOperatorIn,
				Values:   finalIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(finalIds) * 10)
		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		files, _ := client.GetFileBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, file := range files.GetFiles() {
			finalId := file.GetTableId()
			result[finalId] = append(result[finalId], convert.PbFileToModel(file))
		}

		return result, nil
	}
}

// createTopicCouncilsByTopicIdBatchFunc creates a batch function for loading topic councils by topic ID
func createTopicCouncilsByTopicIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.TopicCouncil] {
	return func(ctx context.Context, topicIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.TopicCouncil, error) {
		result := make(map[string][]*model.TopicCouncil)

		if len(topicIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "topic_code",
				Operator: model.FilterOperatorIn,
				Values:   topicIds,
			},
		})

		page := int32(1)

		pageSize := int32(len(topicIds) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		topics, _ := client.GetTopicCouncilBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))

		fmt.Print("cccc", topics)
		for _, topicCouncil := range topics.GetTopicCouncils() {
			topicId := topicCouncil.GetTopicCode()
			result[topicId] = append(result[topicId], convert2.PbTopicCouncilToModel(topicCouncil))
		}

		return result, nil
	}
}

// createEnrollmentsByTopicCouncilIdBatchFunc creates a batch function for loading enrollments by topic council ID
func createEnrollmentsByTopicCouncilIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.Enrollment] {
	return func(ctx context.Context, topicCouncilIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Enrollment, error) {
		result := make(map[string][]*model.Enrollment)

		if len(topicCouncilIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "topic_council_code",
				Operator: model.FilterOperatorIn,
				Values:   topicCouncilIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(topicCouncilIds) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		enrollments, _ := client.GetEnrollmentBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))
		for _, enrollment := range enrollments.GetEnrollments() {
			topicCouncilId := enrollment.GetTopicCouncilCode()
			result[topicCouncilId] = append(result[topicCouncilId], convert2.PbEnrollmentToModel(enrollment))
		}

		return result, nil

	}
}

// createSupervisorsByTopicCouncilIdBatchFunc creates a batch function for loading supervisors by topic council ID
func createSupervisorsByTopicCouncilIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.TopicCouncilSupervisor] {
	return func(ctx context.Context, topicCouncilIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.TopicCouncilSupervisor, error) {
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

			result[topicCouncilId] = convert2.PbTopicCouncilSupervisorsToModel(supervisors.GetTopicCouncilSupervisors())
		}

		return result, nil
	}
}

// createDefencesByCouncilIdBatchFunc creates a batch function for loading defences by council ID
func createDefencesByCouncilIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.Defence] {
	return func(ctx context.Context, councilIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.Defence, error) {
		result := make(map[string][]*model.Defence)

		if len(councilIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "council_code",
				Operator: model.FilterOperatorIn,
				Values:   councilIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(councilIds) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		defences, _ := client.GetDefencesBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, defence := range defences.GetDefences() {
			councilId := defence.GetCouncilCode()
			result[councilId] = append(result[councilId], convert.PbDefenceToModel(defence))
		}

		return result, nil
	}
}

// createTopicCouncilsByCouncilIdBatchFunc creates a batch function for loading topic councils by council ID
func createTopicCouncilsByCouncilIdBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.TopicCouncil] {
	return func(ctx context.Context, councilIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.TopicCouncil, error) {
		result := make(map[string][]*model.TopicCouncil)

		if len(councilIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "council_code",
				Operator: model.FilterOperatorIn,
				Values:   councilIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(councilIds) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		topics, _ := client.GetTopicCouncilBySearch(ctx, convert2.ConvertSearchRequestToPB(*newSearch))

		for _, topicCouncil := range topics.GetTopicCouncils() {
			councilId := topicCouncil.GetCouncilCode()
			result[councilId] = append(result[councilId], convert2.PbTopicCouncilToModel(topicCouncil))
		}

		return result, nil
	}
}

// createGradeDefencesByEnrollmentIdBatchFunc creates a batch function for loading grade defences by enrollment ID
func createGradeDefencesByEnrollmentIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefence] {
	return func(ctx context.Context, enrollmentIds []string, filters []*model.FilterCriteriaInput) (map[string][]*model.GradeDefence, error) {
		result := make(map[string][]*model.GradeDefence)

		if len(enrollmentIds) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "enrollment_code",
				Operator: model.FilterOperatorIn,
				Values:   enrollmentIds,
			},
		})

		page := int32(1)
		pageSize := int32(len(enrollmentIds) * 10)

		newSearch := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		gradeDefences, _ := client.GetGradeDefenceBySearch(ctx, convert.ConvertSearchRequestToPB(*newSearch))

		for _, gradeDefence := range gradeDefences.GetGradeDefences() {
			enrollmentId := gradeDefence.GetEnrollmentCode()
			result[enrollmentId] = append(result[enrollmentId], convert.PbGradeDefenceToModel(gradeDefence))
		}

		return result, nil
	}
}
