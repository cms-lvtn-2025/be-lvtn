package dataloader

import (
	"thaily/src/server/client"
	"time"

	"thaily/src/server/graph/model"
)

// Loaders holds all dataloaders for the application
// Schema 2: Unified types with field-level RBAC
type Loaders struct {
	// ============================================
	// ACADEMIC LOADERS
	// ============================================
	MajorInfoById       *DataLoader[string, *model.MajorInfo]
	SemesterInfoById    *DataLoader[string, *model.SemesterInfo]
	SemesterById        *DataLoader[string, *model.Semester]
	MajorByFacultyId    *DataLoader[string, []*model.Major]
	TopicByMajorId      *DataLoader[string, []*model.Topic]
	StudentBySemesterId *DataLoader[string, []*model.Student]
	TeacherBySemesterId *DataLoader[string, []*model.Teacher]
	TopicBySemesterId   *DataLoader[string, []*model.Topic]

	// ============================================
	// USER LOADERS
	// ============================================
	TeacherInfoById       *DataLoader[string, *model.TeacherInfo]
	TeacherById           *DataLoader[string, *model.Teacher]
	StudentById           *DataLoader[string, *model.Student]
	EnrollmentByStudentId *DataLoader[string, []*model.Enrollment]
	RolesByTeacherId      *DataLoader[string, []*model.RoleSystem]

	// ============================================
	// COUNCIL LOADERS - Unified types
	// ============================================
	CouncilByID                     *DataLoader[string, *model.Council]
	DefenceById                     *DataLoader[string, *model.Defence]
	DefenceByCouncilId              *DataLoader[string, []*model.Defence]
	GradeDefenceByDefenceId         *DataLoader[string, []*model.GradeDefence]
	GradeDefenceByEnrollmentId      *DataLoader[string, []*model.GradeDefence]
	GradeDefenceCriteriaByDefenceId *DataLoader[string, []*model.GradeDefenceCriterion]

	// ============================================
	// THESIS LOADERS - Unified types
	// ============================================
	TopicByID                   *DataLoader[string, *model.Topic]
	TopicCouncilById            *DataLoader[string, *model.TopicCouncil]
	TopicCouncilByTopicId       *DataLoader[string, []*model.TopicCouncil]
	TopicCouncilByCouncilId     *DataLoader[string, []*model.TopicCouncil]
	EnrollmentById              *DataLoader[string, *model.Enrollment]
	EnrollmentByTopicCouncilId  *DataLoader[string, []*model.Enrollment]
	MidtermByID                 *DataLoader[string, *model.Midterm]
	FinalByID                   *DataLoader[string, *model.Final]
	SupervisorsByTopicCouncilId *DataLoader[string, []*model.TopicCouncilSupervisor]
	FilesByTopicId              *DataLoader[string, []*model.File]
	FilesByMidtermId            *DataLoader[string, []*model.File]
	FilesByFinalId              *DataLoader[string, []*model.File]
}

// ============================================
// CACHE INVALIDATION HELPERS
// ============================================

// InvalidateTeacher clears all caches related to a teacher
func (l *Loaders) InvalidateTeacher(id string) {
	l.TeacherById.ClearL2Key(id)
	l.TeacherInfoById.ClearL2Key(id)
	l.RolesByTeacherId.ClearL2Key(id)
}

// InvalidateTeacherBySemester clears teacher list cache for a semester
func (l *Loaders) InvalidateTeacherBySemester(semesterId string) {
	l.TeacherBySemesterId.ClearL2Key(semesterId)
}

// InvalidateStudent clears all caches related to a student
func (l *Loaders) InvalidateStudent(id string) {
	l.StudentById.ClearL2Key(id)
	l.EnrollmentByStudentId.ClearL2Key(id)
}

// InvalidateStudentBySemester clears student list cache for a semester
func (l *Loaders) InvalidateStudentBySemester(semesterId string) {
	l.StudentBySemesterId.ClearL2Key(semesterId)
}

// InvalidateSemester clears all caches related to a semester
func (l *Loaders) InvalidateSemester(id string) {
	l.SemesterById.ClearL2Key(id)
	l.SemesterInfoById.ClearL2Key(id)
	l.StudentBySemesterId.ClearL2Key(id)
	l.TeacherBySemesterId.ClearL2Key(id)
	l.TopicBySemesterId.ClearL2Key(id)
}

// InvalidateMajor clears all caches related to a major
func (l *Loaders) InvalidateMajor(id string) {
	l.MajorInfoById.ClearL2Key(id)
	l.TopicByMajorId.ClearL2Key(id)
}

// InvalidateMajorByFaculty clears major list cache for a faculty
func (l *Loaders) InvalidateMajorByFaculty(facultyId string) {
	l.MajorByFacultyId.ClearL2Key(facultyId)
}

// InvalidateCouncil clears all caches related to a council
func (l *Loaders) InvalidateCouncil(id string) {
	l.CouncilByID.ClearL2Key(id)
	l.DefenceByCouncilId.ClearL2Key(id)
	l.TopicCouncilByCouncilId.ClearL2Key(id)
}

// InvalidateDefence clears all caches related to a defence
func (l *Loaders) InvalidateDefence(id string, councilId string) {
	l.DefenceById.ClearL2Key(id)
	if councilId != "" {
		l.DefenceByCouncilId.ClearL2Key(councilId)
	}
}

// InvalidateTopic clears all caches related to a topic
func (l *Loaders) InvalidateTopic(id string, majorId string, semesterId string) {
	l.TopicByID.ClearL2Key(id)
	l.TopicCouncilByTopicId.ClearL2Key(id)
	if majorId != "" {
		l.TopicByMajorId.ClearL2Key(majorId)
	}
	if semesterId != "" {
		l.TopicBySemesterId.ClearL2Key(semesterId)
	}
}

// InvalidateTopicCouncil clears all caches related to a topic council
func (l *Loaders) InvalidateTopicCouncil(id string, topicId string, councilId string) {
	l.TopicCouncilById.ClearL2Key(id)
	l.EnrollmentByTopicCouncilId.ClearL2Key(id)
	l.SupervisorsByTopicCouncilId.ClearL2Key(id)
	if topicId != "" {
		l.TopicCouncilByTopicId.ClearL2Key(topicId)
	}
	if councilId != "" {
		l.TopicCouncilByCouncilId.ClearL2Key(councilId)
	}
}

// InvalidateEnrollment clears all caches related to an enrollment
func (l *Loaders) InvalidateEnrollment(id string, studentId string, topicCouncilId string) {
	l.EnrollmentById.ClearL2Key(id)
	if studentId != "" {
		l.EnrollmentByStudentId.ClearL2Key(studentId)
	}
	if topicCouncilId != "" {
		l.EnrollmentByTopicCouncilId.ClearL2Key(topicCouncilId)
	}
}

// InvalidateGradeDefence clears all caches related to a grade defence
func (l *Loaders) InvalidateGradeDefence(defenceId string, enrollmentId string) {
	if defenceId != "" {
		l.GradeDefenceByDefenceId.ClearL2Key(defenceId)
		l.GradeDefenceCriteriaByDefenceId.ClearL2Key(defenceId)
	}
	if enrollmentId != "" {
		l.GradeDefenceByEnrollmentId.ClearL2Key(enrollmentId)
	}
}

// InvalidateMidterm clears all caches related to a midterm
func (l *Loaders) InvalidateMidterm(id string) {
	l.MidtermByID.ClearL2Key(id)
	l.FilesByMidtermId.ClearL2Key(id)
}

// InvalidateFinal clears all caches related to a final
func (l *Loaders) InvalidateFinal(id string) {
	l.FinalByID.ClearL2Key(id)
	l.FilesByFinalId.ClearL2Key(id)
}

// InvalidateFilesByTopic clears file cache for a topic
func (l *Loaders) InvalidateFilesByTopic(topicId string) {
	l.FilesByTopicId.ClearL2Key(topicId)
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
	// L2TTL = 0 disables persistent cache, each request batch gets fresh data
	defaultConfig := &Config{
		BatchWindow:  2 * time.Millisecond,
		MaxBatchSize: 300,
		L2TTL:        0, // Disable L2 cache - only batch within same request
	}

	return &Loaders{
		// ============================================
		// ACADEMIC LOADERS
		// ============================================
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

		// ============================================
		// USER LOADERS
		// ============================================
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

		// ============================================
		// COUNCIL LOADERS
		// ============================================
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
		GradeDefenceByDefenceId: NewDataLoader(
			createGradeDefenceByDefenceIdBatchFunc(councilClient),
			defaultConfig,
		),
		GradeDefenceByEnrollmentId: NewDataLoader(
			createGradeDefencesByEnrollmentIdBatchFunc(councilClient),
			defaultConfig,
		),
		GradeDefenceCriteriaByDefenceId: NewDataLoader(
			createGradeDefenceCriteriaByDefenceIdBatchFunc(councilClient),
			defaultConfig,
		),

		// ============================================
		// THESIS LOADERS
		// ============================================
		TopicByID: NewDataLoader(
			createTopicBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicCouncilById: NewDataLoader(
			createTopicCouncilByIdBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicCouncilByTopicId: NewDataLoader(
			createTopicCouncilsByTopicIdBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicCouncilByCouncilId: NewDataLoader(
			createTopicCouncilsByCouncilIdBatchFunc(thesisClient),
			defaultConfig,
		),
		EnrollmentById: NewDataLoader(
			createEnrollmentByIdBatchFunc(thesisClient),
			defaultConfig,
		),
		EnrollmentByTopicCouncilId: NewDataLoader(
			createEnrollmentsByTopicCouncilIdBatchFunc(thesisClient),
			defaultConfig,
		),
		MidtermByID: NewDataLoader(
			createMidtermBatchFunc(thesisClient),
			defaultConfig,
		),
		FinalByID: NewDataLoader(
			createFinalBatchFunc(thesisClient),
			defaultConfig,
		),
		SupervisorsByTopicCouncilId: NewDataLoader(
			createSupervisorsByTopicCouncilIdBatchFunc(thesisClient),
			defaultConfig,
		),
		FilesByTopicId: NewDataLoader(
			createFilesByTopicIdBatchFunc(fileClient),
			defaultConfig,
		),
		FilesByMidtermId: NewDataLoader(
			createFilesByMidtermIdBatchFunc(fileClient),
			defaultConfig,
		),
		FilesByFinalId: NewDataLoader(
			createFilesByFinalIdBatchFunc(fileClient),
			defaultConfig,
		),
	}
}
