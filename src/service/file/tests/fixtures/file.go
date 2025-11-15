package fixtures

import (
	pb "thaily/proto/file"
)

// File test IDs
const (
	FileID1 = "file-uuid-1111-1111-111111111111"
	FileID2 = "file-uuid-2222-2222-222222222222"
)

// File test data
const (
	FileTitle1   = "Test File 1"
	FileTitle2   = "Test File 2"
	FileName1    = "file1.pdf"
	FileName2    = "file2.pdf"
	FileOption1  = "option1"
	FileOption2  = "option2"
	FileTableID1 = "table-id-1"
	FileTableID2 = "table-id-2"
	CreatedBy    = "user-123"
	UpdatedBy    = "user-123"
)

// File1 is a test File entity with FILE_PENDING status and TOPIC table
var File1 = &pb.File{
	Id:        FileID1,
	Title:     FileTitle1,
	File:      FileName1,
	Status:    pb.FileStatus_FILE_PENDING,
	Table:     pb.TableType_TOPIC,
	Option:    FileOption1,
	TableId:   FileTableID1,
	CreatedBy: CreatedBy,
	UpdatedBy: UpdatedBy,
}

// File2 is a test File entity with APPROVED status and MIDTERM table
var File2 = &pb.File{
	Id:        FileID2,
	Title:     FileTitle2,
	File:      FileName2,
	Status:    pb.FileStatus_APPROVED,
	Table:     pb.TableType_MIDTERM,
	Option:    FileOption2,
	TableId:   FileTableID2,
	CreatedBy: CreatedBy,
	UpdatedBy: UpdatedBy,
}
