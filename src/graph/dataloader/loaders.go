package dataloader

import (
	"time"

	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// Loaders holds all dataloaders for the application
type Loaders struct {
	// Academic - General
	MajorInfoById    *DataLoader[string, *model.MajorInfo]
	SemesterInfoById *DataLoader[string, *model.SemesterInfo]
	SemesterById     *DataLoader[string, *model.Semester]
	// Academic - Relationships
	MajorByFacultyId      *DataLoader[string, []*model.Major]
	TopicByMajorId        *DataLoader[string, []*model.Topic]
	StudentBySemesterId   *DataLoader[string, []*model.Student]
	TeacherBySemesterId   *DataLoader[string, []*model.Teacher]
	TopicBySemesterId     *DataLoader[string, []*model.Topic]
	// Council - Student views
	DefenceInfoByCouncilId          *DataLoader[string, []*model.StudentDefenceInfo]
	CouncilByIdForStudent           *DataLoader[string, *model.StudentCouncil]
	GradeDefencesInfoByEnrollmentId *DataLoader[string, []*model.StudentGradeDefence]
	GradeDefenceCriteriaByDefenceId *DataLoader[string, []*model.GradeDefenceCriterion]
	DefenceInfoById                 *DataLoader[string, *model.StudentDefenceInfo]
	// Council - Teacher views
	CouncilMemberById       *DataLoader[string, *model.CouncilMemberCouncil]
	GradeDefenceByDefenceId *DataLoader[string, []*model.GradeDefence]
	CouncilDefenceById      *DataLoader[string, *model.CouncilDefence]
	// Council - General
	CouncilByID                *DataLoader[string, *model.Council]
	DefenceById                *DataLoader[string, *model.Defence]
	DefenceByCouncilId         *DataLoader[string, []*model.Defence]
	TopicCouncilByCouncilId    *DataLoader[string, []*model.TopicCouncil]
	GradeDefenceByEnrollmentId *DataLoader[string, []*model.GradeDefence]
	// User
	TeacherInfoById     *DataLoader[string, *model.StudentTeacherInfo]
	TeacherById         *DataLoader[string, *model.Teacher]
	StudentById         *DataLoader[string, *model.Student]
	EnrollmentByStudentId *DataLoader[string, []*model.Enrollment]
	RolesByTeacherId      *DataLoader[string, []*model.RoleSystem]
	// Thesis - General
	TopicCouncilInfoById           *DataLoader[string, *model.StudentTopicCouncil]
	MidtermByID                    *DataLoader[string, *model.Midterm]
	FinalByID                      *DataLoader[string, *model.Final]
	GradeViewById                  *DataLoader[string, *model.GradeReview]
	TopicByID                      *DataLoader[string, *model.Topic]
	TopicForStudentByID            *DataLoader[string, *model.StudentTopic]
	SupervisorByTopicCouncilId     *DataLoader[string, []*model.StudentTopicSupervisor]
	EnrollmentById                 *DataLoader[string, *model.Enrollment]
	TopicCouncilById               *DataLoader[string, *model.TopicCouncil]
	// Thesis - Relationships
	FilesByTopicId                 *DataLoader[string, []*model.File]
	TopicCouncilByTopicId          *DataLoader[string, []*model.TopicCouncil]
	EnrollmentByTopicCouncilId     *DataLoader[string, []*model.Enrollment]
	SupervisorsByTopicCouncilId    *DataLoader[string, []*model.TopicCouncilSupervisor]
	// Teacher - Thesis
	CouncilTopicCouncilById    *DataLoader[string, *model.CouncilTopicCouncil]
	ReviewerTopicCouncilById   *DataLoader[string, *model.ReviewerTopicCouncil]
	SupervisorTopicCouncilById *DataLoader[string, *model.SupervisorTopicCouncil]
	ReviewerTopicById          *DataLoader[string, *model.ReviewerTopic]
	SupervisorTopicById        *DataLoader[string, *model.SupervisorTopic]
}

// NewLoaders creates a new Loaders instance with all dataloaders
func NewLoaders(
	userClient *client.GRPCUser,
	thesisClient *client.GRPCthesis,
	councilClient *client.GRPCCouncil,
	academicClient *client.GRPCAcadamicClient,
	roleClient *client.GRPCRole,
	fileClient *client.GRPCfile,
) *Loaders {
	// Default configuration for all loaders
	defaultConfig := &Config{
		BatchWindow:  2 * time.Millisecond,
		MaxBatchSize: 300,
		L2TTL:        5 * time.Minute,
	}

	return &Loaders{
		// Academic - General
		MajorInfoById: NewDataLoader(
			createMajorInfoBatchFunc(academicClient),
			defaultConfig,
		),
		SemesterInfoById: NewDataLoader(
			createSemesterInfoBatchFunc(academicClient),
			defaultConfig,
		),
		SemesterById: NewDataLoader(
			createSemesterByIdBatchFunc(academicClient),
			defaultConfig,
		),
		// Academic - Relationships
		MajorByFacultyId: NewDataLoader(
			createMajorsByFacultyIdBatchFunc(academicClient),
			defaultConfig,
		),
		TopicByMajorId: NewDataLoader(
			createTopicsByMajorIdBatchFunc(thesisClient),
			defaultConfig,
		),
		StudentBySemesterId: NewDataLoader(
			createStudentsBySemesterIdBatchFunc(userClient),
			defaultConfig,
		),
		TeacherBySemesterId: NewDataLoader(
			createTeachersBySemesterIdBatchFunc(userClient),
			defaultConfig,
		),
		TopicBySemesterId: NewDataLoader(
			createTopicsBySemesterIdBatchFunc(thesisClient),
			defaultConfig,
		),
		// Council - Student views
		DefenceInfoByCouncilId: NewDataLoader(
			createDefenceByCouncilIDBatchFunc(councilClient),
			defaultConfig,
		),
		DefenceInfoById: NewDataLoader(
			createDefenceInfoByIdBatchFunc(councilClient),
			defaultConfig,
		),
		CouncilByIdForStudent: NewDataLoader(
			createCouncilForStudentBatchFunc(councilClient),
			defaultConfig,
		),
		GradeDefencesInfoByEnrollmentId: NewDataLoader(
			createGradeDefencesForStudentBatchFunc(councilClient),
			defaultConfig,
		),
		GradeDefenceCriteriaByDefenceId: NewDataLoader(
			createGradeDefenceCriteriaForByDefenceIdBatchFunc(councilClient),
			defaultConfig,
		),
		// Council - Teacher views
		CouncilMemberById: NewDataLoader(
			createCouncilMemberCouncilBatchFunc(councilClient),
			defaultConfig,
		),
		GradeDefenceByDefenceId: NewDataLoader(
			createGradeDefenceByDefenceIdBatchFunc(councilClient),
			defaultConfig,
		),
		CouncilDefenceById: NewDataLoader(
			createCouncilDefenceBatchFunc(councilClient),
			defaultConfig,
		),
		// Council - General
		CouncilByID: NewDataLoader(
			createCouncilByIdBatchFunc(councilClient),
			defaultConfig,
		),
		DefenceById: NewDataLoader(
			createDefenceByIdBatchFunc(councilClient),
			defaultConfig,
		),
		DefenceByCouncilId: NewDataLoader(
			createDefencesByCouncilIdBatchFunc(councilClient),
			defaultConfig,
		),
		TopicCouncilByCouncilId: NewDataLoader(
			createTopicCouncilsByCouncilIdBatchFunc(thesisClient),
			defaultConfig,
		),
		GradeDefenceByEnrollmentId: NewDataLoader(
			createGradeDefencesByEnrollmentIdBatchFunc(councilClient),
			defaultConfig,
		),
		// User
		TeacherInfoById: NewDataLoader(
			createTeacherInfoBatchFunc(userClient),
			defaultConfig,
		),
		TeacherById: NewDataLoader(
			createTeacherBatchFunc(userClient),
			defaultConfig,
		),
		StudentById: NewDataLoader(
			createStudentByIdBatchFunc(userClient),
			defaultConfig,
		),
		EnrollmentByStudentId: NewDataLoader(
			createEnrollmentsByStudentIdBatchFunc(thesisClient),
			defaultConfig,
		),
		RolesByTeacherId: NewDataLoader(
			createRolesByTeacherIdBatchFunc(roleClient),
			defaultConfig,
		),
		// Thesis - General
		TopicCouncilInfoById: NewDataLoader(
			createStudentTopicCouncilInfoBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicForStudentByID: NewDataLoader(
			createTopicForStudentBatchFunc(thesisClient),
			defaultConfig,
		),
		MidtermByID: NewDataLoader(
			createMidtermBatchFunc(thesisClient),
			defaultConfig,
		),
		SupervisorByTopicCouncilId: NewDataLoader(
			createSupervisorForStudentBatchFunc(thesisClient),
			defaultConfig,
		),
		GradeViewById: NewDataLoader(
			createGradeViewBatchFunc(thesisClient),
			defaultConfig,
		),
		FinalByID: NewDataLoader(
			createFinalBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicByID: NewDataLoader(
			createTopicBatchFunc(thesisClient),
			defaultConfig,
		),
		EnrollmentById: NewDataLoader(
			createEnrollmentByIdBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicCouncilById: NewDataLoader(
			createTopicCouncilByIdBatchFunc(thesisClient),
			defaultConfig,
		),
		// Thesis - Relationships
		FilesByTopicId: NewDataLoader(
			createFilesByTopicIdBatchFunc(fileClient),
			defaultConfig,
		),
		TopicCouncilByTopicId: NewDataLoader(
			createTopicCouncilsByTopicIdBatchFunc(thesisClient),
			defaultConfig,
		),
		EnrollmentByTopicCouncilId: NewDataLoader(
			createEnrollmentsByTopicCouncilIdBatchFunc(thesisClient),
			defaultConfig,
		),
		SupervisorsByTopicCouncilId: NewDataLoader(
			createSupervisorsByTopicCouncilIdBatchFunc(thesisClient),
			defaultConfig,
		),
		// Teacher - Thesis
		CouncilTopicCouncilById: NewDataLoader(
			createCouncilTopicCouncilBatchFunc(thesisClient),
			defaultConfig,
		),
		ReviewerTopicCouncilById: NewDataLoader(
			createReviewerTopicCouncilBatchFunc(thesisClient),
			defaultConfig,
		),
		SupervisorTopicCouncilById: NewDataLoader(
			createSupervisorTopicCouncilBatchFunc(thesisClient),
			defaultConfig,
		),
		ReviewerTopicById: NewDataLoader(
			createReviewerTopicBatchFunc(thesisClient),
			defaultConfig,
		),
		SupervisorTopicById: NewDataLoader(
			createSupervisorTopicBatchFunc(thesisClient),
			defaultConfig,
		),
	}
}
