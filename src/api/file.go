package api

import (
	"thaily/src/pkg/response"

	"github.com/gin-gonic/gin"
)

type FileUpload struct {
}

// UploadFile xử lý upload file
func (h *APIHandler) UploadFile(c *gin.Context) {
	if h.FileClient == nil {
		response.InternalError(c, "File service not available")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded")
		return
	}

	// TODO: Implement upload logic với FileClient
	// 1. Đọc file content
	// 2. Gọi FileClient để upload
	// 3. Return file URL

	response.SuccessWithMessage(c, "File uploaded successfully", gin.H{
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// GetFile lấy thông tin file
func (h *APIHandler) GetFile(c *gin.Context) {
	if h.FileClient == nil {
		response.InternalError(c, "File service not available")
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	// TODO: Gọi FileClient để lấy file info

	response.Success(c, gin.H{
		"id": fileID,
	})
}
