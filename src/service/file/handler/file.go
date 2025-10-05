package handler

import (
	"context"
	pb "thaily/proto/file"
)


func (h *Handler) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	// TODO: Implement CreateFile
	return &pb.CreateFileResponse{}, nil
}

func (h *Handler) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	// TODO: Implement GetFile
	return &pb.GetFileResponse{}, nil
}

func (h *Handler) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileResponse, error) {
	// TODO: Implement UpdateFile
	return &pb.UpdateFileResponse{}, nil
}

func (h *Handler) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	// TODO: Implement DeleteFile
	return &pb.DeleteFileResponse{}, nil
}

func (h *Handler) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	// TODO: Implement ListFiles
	return &pb.ListFilesResponse{}, nil
}

