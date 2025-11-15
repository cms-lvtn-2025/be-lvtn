package fixtures

import pb "thaily/proto/academic"

// SemesterTestData provides test data for Semester operations
var SemesterTestData = struct {
	ValidSemester   *pb.Semester
	InvalidSemester *pb.Semester
	Semesters       []*pb.Semester
}{
	ValidSemester: &pb.Semester{
		Id:        "semester-1",
		Title:     "Fall 2024",
		CreatedBy: "admin",
	},
	InvalidSemester: &pb.Semester{
		Id:        "",
		Title:     "",
		CreatedBy: "",
	},
	Semesters: []*pb.Semester{
		{
			Id:        "semester-1",
			Title:     "Fall 2024",
			CreatedBy: "admin",
		},
		{
			Id:        "semester-2",
			Title:     "Spring 2024",
			CreatedBy: "admin",
		},
		{
			Id:        "semester-3",
			Title:     "Summer 2024",
			CreatedBy: "admin",
		},
	},
}

// CreateSemesterRequests provides various request scenarios for testing
var CreateSemesterRequests = struct {
	Valid   *pb.CreateSemesterRequest
	Empty   *pb.CreateSemesterRequest
	NoTitle *pb.CreateSemesterRequest
}{
	Valid: &pb.CreateSemesterRequest{
		Title:     "Fall 2024",
		CreatedBy: "admin",
	},
	Empty: &pb.CreateSemesterRequest{
		Title:     "",
		CreatedBy: "",
	},
	NoTitle: &pb.CreateSemesterRequest{
		Title:     "",
		CreatedBy: "admin",
	},
}

// UpdateSemesterRequests provides update scenarios for testing
var UpdateSemesterRequests = struct {
	Valid   *pb.UpdateSemesterRequest
	Invalid *pb.UpdateSemesterRequest
}{
	Valid: &pb.UpdateSemesterRequest{
		Id:        "semester-1",
		Title:     &[]string{"Updated Fall 2024"}[0],
		UpdatedBy: "admin",
	},
	Invalid: &pb.UpdateSemesterRequest{
		Id:        "",
		Title:     nil,
		UpdatedBy: "",
	},
}

// CommonTestData provides data used across different entity types
var CommonTestData = struct {
	AdminUser     string
	TestUser      string
	InvalidID     string
	NonExistentID string
}{
	AdminUser:     "admin",
	TestUser:      "test-user",
	InvalidID:     "",
	NonExistentID: "non-existent-999",
}
