package controller

import (
	"context"
	"fmt"

	pbFile "thaily/proto/file"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/dataloader"
	"thaily/src/server/graph/model"
)

// ============================================
// STUDENT MUTATION CONTROLLER
// Handles: StudentMutation resolvers
// ============================================

// UploadMidtermFile uploads a midterm file for the current student
func (c *Controller) UploadMidtermFile(ctx context.Context, input model.UploadFileInput) (*model.File, error) {
	myId, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if myId == nil || role == nil || *role != "student" {
		return nil, fmt.Errorf("not authorized: only students can upload midterm files")
	}

	option := ""
	if input.Option != nil {
		option = *input.Option
	}

	resp, err := c.file.CreateFile(ctx, &pbFile.CreateFileRequest{
		File: &pbFile.FileAction{
			Title:     input.Title,
			File:      input.File,
			Status:    pbFile.FileStatus_FILE_PENDING,
			Table:     pbFile.TableType_MIDTERM,
			TableId:   input.TableID,
			Option:    option,
			CreatedBy: *myId,
		},
	})
	if err != nil {
		return nil, err
	}

	// Invalidate cache - file is associated with midterm via TableID
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateMidterm(input.TableID)
	}

	return convert.PbFileToModel(resp.GetFile()), nil
}

// UploadFinalFile uploads a final file for the current student
func (c *Controller) UploadFinalFile(ctx context.Context, input model.UploadFileInput) (*model.File, error) {
	myId, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if myId == nil || role == nil || *role != "student" {
		return nil, fmt.Errorf("not authorized: only students can upload final files")
	}

	option := ""
	if input.Option != nil {
		option = *input.Option
	}

	resp, err := c.file.CreateFile(ctx, &pbFile.CreateFileRequest{
		File: &pbFile.FileAction{
			Title:     input.Title,
			File:      input.File,
			Status:    pbFile.FileStatus_FILE_PENDING,
			Table:     pbFile.TableType_FINAL,
			TableId:   input.TableID,
			Option:    option,
			CreatedBy: *myId,
		},
	})
	if err != nil {
		return nil, err
	}

	// Invalidate cache - file is associated with final via TableID
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateFinal(input.TableID)
	}

	return convert.PbFileToModel(resp.GetFile()), nil
}
