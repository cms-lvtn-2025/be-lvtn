package dataloader

import (
	"time"

	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// Loaders holds all dataloaders for the application
type Loaders struct {
	// Academic
	MajorInfoById    *DataLoader[string, *model.MajorInfo]
	SemesterInfoById *DataLoader[string, *model.SemesterInfo]
	// Council
	DefenceInfoByCouncilId          *DataLoader[string, []*model.StudentDefenceInfo]
	CouncilByIdForStudent           *DataLoader[string, *model.StudentCouncil]
	GradeDefencesInfoByEnrollmentId *DataLoader[string, []*model.StudentGradeDefence]
	GradeDefenceCriteriaByDefenceId *DataLoader[string, []*model.GradeDefenceCriterion]
	DefenceInfoById                 *DataLoader[string, *model.StudentDefenceInfo]
	// User
	TeacherInfoById *DataLoader[string, *model.StudentTeacherInfo]
	// Thesis
	TopicCouncilInfoById       *DataLoader[string, *model.StudentTopicCouncil]
	CouncilByID                *DataLoader[string, *model.Council]
	MidtermByID                *DataLoader[string, *model.Midterm]
	FinalByID                  *DataLoader[string, *model.Final]
	GradeViewById              *DataLoader[string, *model.GradeReview]
	TopicByID                  *DataLoader[string, *model.Topic]
	TopicForStudentByID        *DataLoader[string, *model.StudentTopic]
	SupervisorByTopicCouncilId *DataLoader[string, []*model.StudentTopicSupervisor]
}

// NewLoaders creates a new Loaders instance with all dataloaders
func NewLoaders(
	userClient *client.GRPCUser,
	thesisClient *client.GRPCthesis,
	councilClient *client.GRPCCouncil,
	academic *client.GRPCAcadamicClient,
) *Loaders {
	// Default configuration for all loaders
	defaultConfig := &Config{
		BatchWindow:  2 * time.Millisecond,
		MaxBatchSize: 300,
		L2TTL:        5 * time.Minute,
	}

	return &Loaders{
		// academic
		MajorInfoById: NewDataLoader(
			createMajorInfoBatchFunc(academic),
			defaultConfig,
		),
		SemesterInfoById: NewDataLoader(
			createSemesterInfoBatchFunc(academic),
			defaultConfig,
		),
		// council
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

		// User
		TeacherInfoById: NewDataLoader(
			createTeacherInfoBatchFunc(userClient),
			defaultConfig,
		),

		// Thesis
		TopicCouncilInfoById: NewDataLoader(
			createStudentTopicCouncilInfoBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicForStudentByID: NewDataLoader(
			createTopicForStudentBatchFunc(thesisClient), defaultConfig),

		MidtermByID: NewDataLoader(
			createMidtermBatchFunc(thesisClient),
			defaultConfig,
		),

		SupervisorByTopicCouncilId: NewDataLoader(
			createSupervisorForStudentBatchFunc(thesisClient), defaultConfig,
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
	}
}
