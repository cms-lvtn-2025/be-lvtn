package fixtures

import pb "thaily/proto/academic"

// FacultyTestData provides test data for Faculty operations
var FacultyTestData = struct {
	ValidFaculty   *pb.Faculty
	InvalidFaculty *pb.Faculty
	Faculties      []*pb.Faculty
}{
	ValidFaculty: &pb.Faculty{
		Id:        "faculty-1",
		Title:     "Computer Science",
		CreatedBy: "admin",
	},
	InvalidFaculty: &pb.Faculty{
		Id:        "",
		Title:     "",
		CreatedBy: "",
	},
	Faculties: []*pb.Faculty{
		{
			Id:        "faculty-1",
			Title:     "Computer Science",
			CreatedBy: "admin",
		},
		{
			Id:        "faculty-2",
			Title:     "Mathematics",
			CreatedBy: "admin",
		},
		{
			Id:        "faculty-3",
			Title:     "Physics",
			CreatedBy: "admin",
		},
	},
}

// CreateFacultyRequests provides various request scenarios for testing
var CreateFacultyRequests = struct {
	Valid   *pb.CreateFacultyRequest
	Empty   *pb.CreateFacultyRequest
	NoTitle *pb.CreateFacultyRequest
}{
	Valid: &pb.CreateFacultyRequest{
		Title:     "Computer Science",
		CreatedBy: "admin",
	},
	Empty: &pb.CreateFacultyRequest{
		Title:     "",
		CreatedBy: "",
	},
	NoTitle: &pb.CreateFacultyRequest{
		Title:     "",
		CreatedBy: "admin",
	},
}

// UpdateFacultyRequests provides update scenarios for testing
var UpdateFacultyRequests = struct {
	Valid   *pb.UpdateFacultyRequest
	Invalid *pb.UpdateFacultyRequest
}{
	Valid: &pb.UpdateFacultyRequest{
		Id:        "faculty-1",
		Title:     &[]string{"Updated Computer Science"}[0],
		UpdatedBy: "admin",
	},
	Invalid: &pb.UpdateFacultyRequest{
		Id:        "",
		Title:     nil,
		UpdatedBy: "",
	},
}
