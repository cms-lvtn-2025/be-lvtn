package convert

import (
	pbFile "thaily/proto/file"
	"thaily/src/graph/model"
)

// PbFileToModel converts protobuf File to GraphQL File
func PbFileToModel(pb *pbFile.File) *model.File {
	if pb == nil {
		return nil
	}

	result := &model.File{
		ID:      pb.Id,
		Title:   pb.Title,
		Status:  PbFileStatusToModel(pb.Status),
		Table:   PbTableTypeToModel(pb.Table),
		TableID: pb.TableId,
	}

	// Handle optional File field
	if pb.File != "" {
		result.File = &pb.File
	}

	// Handle optional Option
	if pb.Option != "" {
		result.Option = &pb.Option
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbFileStatusToModel converts protobuf FileStatus enum to GraphQL FileStatus enum
func PbFileStatusToModel(pb pbFile.FileStatus) model.FileStatus {
	switch pb {
	case pbFile.FileStatus_FILE_PENDING:
		return model.FileStatusFilePending
	case pbFile.FileStatus_APPROVED:
		return model.FileStatusApproved
	case pbFile.FileStatus_REJECTED:
		return model.FileStatusRejected
	default:
		return model.FileStatusFilePending
	}
}

// PbTableTypeToModel converts protobuf TableType enum to GraphQL FileTable enum
func PbTableTypeToModel(pb pbFile.TableType) model.FileTable {
	switch pb {
	case pbFile.TableType_TOPIC:
		return model.FileTableTopic
	case pbFile.TableType_MIDTERM:
		return model.FileTableMidterm
	case pbFile.TableType_FINAL:
		return model.FileTableFinal
	case pbFile.TableType_ORDER:
		return model.FileTableOrder
	default:
		return model.FileTableTopic
	}
}

// PbFilesToModel converts array of protobuf Files to GraphQL Files
func PbFilesToModel(pbs []*pbFile.File) []*model.File {
	if pbs == nil {
		return nil
	}

	result := make([]*model.File, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbFileToModel(pb))
		}
	}
	return result
}

// ============================================
// LIST RESPONSE FACTORY FUNCTIONS
// ============================================

// CreateFileListResponse creates a FileListResponse
func CreateFileListResponse(files []*model.File, total int32) *model.FileListResponse {
	return &model.FileListResponse{
		Data:  files,
		Total: total,
	}
}
