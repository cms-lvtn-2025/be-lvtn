package fixtures

import pb "thaily/proto/academic"

// MajorTestData provides test data for Major operations
var MajorTestData = struct {
	ValidMajor   *pb.Major
	InvalidMajor *pb.Major
	Majors       []*pb.Major
}{
	ValidMajor: &pb.Major{
		Id:          "major-1",
		Title:       "Computer Science",
		FacultyCode: "CSC",
		CreatedBy:   "admin",
	},
	InvalidMajor: &pb.Major{
		Id:          "",
		Title:       "",
		FacultyCode: "",
		CreatedBy:   "",
	},
	Majors: []*pb.Major{
		{
			Id:          "major-1",
			Title:       "Computer Science",
			FacultyCode: "CSC",
			CreatedBy:   "admin",
		},
		{
			Id:          "major-2",
			Title:       "Mathematics",
			FacultyCode: "MATH",
			CreatedBy:   "admin",
		},
		{
			Id:          "major-3",
			Title:       "Software Engineering",
			FacultyCode: "CSC",
			CreatedBy:   "admin",
		},
	},
}

// CreateMajorRequests provides various request scenarios for testing
var CreateMajorRequests = struct {
	Valid         *pb.CreateMajorRequest
	Empty         *pb.CreateMajorRequest
	NoTitle       *pb.CreateMajorRequest
	NoFacultyCode *pb.CreateMajorRequest
}{
	Valid: &pb.CreateMajorRequest{
		Title:       "Computer Science",
		FacultyCode: "CSC",
		CreatedBy:   "admin",
	},
	Empty: &pb.CreateMajorRequest{
		Title:       "",
		FacultyCode: "",
		CreatedBy:   "",
	},
	NoTitle: &pb.CreateMajorRequest{
		Title:       "",
		FacultyCode: "CSC",
		CreatedBy:   "admin",
	},
	NoFacultyCode: &pb.CreateMajorRequest{
		Title:       "Computer Science",
		FacultyCode: "",
		CreatedBy:   "admin",
	},
}

// UpdateMajorRequests provides update scenarios for testing
var UpdateMajorRequests = struct {
	Valid   *pb.UpdateMajorRequest
	Invalid *pb.UpdateMajorRequest
}{
	Valid: &pb.UpdateMajorRequest{
		Id:          "major-1",
		Title:       &[]string{"Updated Computer Science"}[0],
		FacultyCode: &[]string{"CSC2"}[0],
		UpdatedBy:   "admin",
	},
	Invalid: &pb.UpdateMajorRequest{
		Id:          "",
		Title:       nil,
		FacultyCode: nil,
		UpdatedBy:   "",
	},
}
