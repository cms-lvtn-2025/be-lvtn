package handler

import (
	"github.com/gin-gonic/gin"

	pbRole "thaily/proto/role"
	"thaily/src/server/api/model"
	"thaily/src/server/api/service"
	"thaily/src/server/response"
)

// uploadFileHandler is the base handler for file uploads
func (h *Handler) uploadFileHandler(c *gin.Context, uploadType model.FileUploadType, allowedRoles []string) {
	// Extract user info
	userInfo, err := h.AuthService.ExtractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Check role permission
	roleAllowed, err := h.FileService.CheckRolePermission(c.Request.Context(), userInfo, allowedRoles)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if !roleAllowed {
		response.Forbidden(c, "You don't have permission to upload this file")
		return
	}

	// Check semester for certain upload types
	if uploadType != model.UploadTypeFinal && userInfo.Semester == "" {
		response.BadRequest(c, "semester is required")
		return
	}

	// Get file from request
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "no file uploaded")
		return
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c, "failed to open uploaded file")
		return
	}
	defer file.Close()

	// Get title
	title := c.PostForm("title")
	if title == "" {
		title = fileHeader.Filename
	}

	// Prepare upload params
	params := &service.UploadFileParams{
		File:         file,
		FileHeader:   fileHeader,
		UploadType:   uploadType,
		UserInfo:     userInfo,
		Title:        title,
		TableType:    c.PostForm("table_type"),
		TableID:      c.PostForm("table_id"),
		EnrollmentID: c.PostForm("enrollment_id"),
		SemesterCode: c.PostForm("semester_code"),
	}

	// Upload file
	result, err := h.FileService.UploadFile(c.Request.Context(), c, params)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "File uploaded successfully", gin.H{
		"file_id":       result.FileID,
		"filename":      result.Filename,
		"size":          result.Size,
		"url":           result.URL,
		"object_path":   result.ObjectPath,
		"uploaded_by":   result.UploadedBy,
		"uploaded_role": result.UploadedRole,
	})
}

// UploadGradeSuppervisorFile handles grade supervisor upload (Excel files)
// POST /api/files/upload/grade-supervisor
func (h *Handler) UploadGradeSuppervisorFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadTypeGradeSuppervisor, []string{"teacher"})
}

// UploadGradeDefenceFile handles grade defence upload (Excel files)
// POST /api/files/upload/grade-defence
func (h *Handler) UploadGradeDefenceFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadTypeGradeDefence, []string{"teacher"})
}

// UploadTopicCouncilForDepartmentFile handles topic council for department upload (Excel files)
// POST /api/files/upload/topic-council-for-department
func (h *Handler) UploadTopicCouncilForDepartmentFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadTopicCouncilForDepartment, []string{"department"})
}

// UploadCouncilForDepartmentFile handles council for department upload (Excel files)
// POST /api/files/upload/council-for-department
func (h *Handler) UploadCouncilForDepartmentFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadCouncilForDepartment, []string{"department"})
}

// UploadCouncilForAffairFile handles council for affair upload (Excel files)
// POST /api/files/upload/council-for-affair
func (h *Handler) UploadCouncilForAffairFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadCouncilForAffair, []string{"affair"})
}

// UploadUserForAffairFile handles user for affair upload (Excel files)
// POST /api/files/upload/user-for-affair
func (h *Handler) UploadUserForAffairFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadStudentForAffair, []string{"affair"})
}

// UploadTeacherForAffairFile handles teacher for affair upload (Excel files)
// POST /api/files/upload/teacher-for-affair
func (h *Handler) UploadTeacherForAffairFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadTeacherForAffair, []string{"affair"})
}

// UploadMidtermFile handles midterm upload (PDF files)
// POST /api/files/upload/midterm
func (h *Handler) UploadMidtermFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadTypeMidterm, []string{"student"})
}

// UploadFinalFile handles final upload (PDF files)
// POST /api/files/upload/final
func (h *Handler) UploadFinalFile(c *gin.Context) {
	h.uploadFileHandler(c, model.UploadTypeFinal, []string{"student"})
}

// GetFileURL generates a presigned URL for file download
// GET /api/files/:id/url
func (h *Handler) GetFileURL(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	userInfo, err := h.AuthService.ExtractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	isExcel := c.Query("excel") == "true"
	urlType := c.Query("type")

	result, err := h.FileService.GetFileURL(c.Request.Context(), c, fileID, userInfo, isExcel, urlType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteFile deletes a file
// DELETE /api/files/:id
func (h *Handler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	userInfo, err := h.AuthService.ExtractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	isExcel := c.Query("excel") == "true"

	if err := h.FileService.DeleteFile(c.Request.Context(), fileID, userInfo, isExcel); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "File deleted successfully", gin.H{
		"file_id": fileID,
	})
}

// ListFilesExcel lists excel files based on role
// GET /api/files/list/excel
func (h *Handler) ListFilesExcel(c *gin.Context) {
	userInfo, err := h.AuthService.ExtractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	var roleHeader pbRole.RoleType
	switch c.GetHeader("x-role") {
	case "affair":
		roleHeader = pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF
	case "department":
		roleHeader = pbRole.RoleType_DEPARTMENT_LECTURER
	case "teacher":
		roleHeader = pbRole.RoleType_TEACHER
	default:
		response.BadRequest(c, "Invalid role type")
		return
	}

	files, err := h.FileService.ListExcelFiles(c.Request.Context(), userInfo, roleHeader)
	if err != nil {
		response.Forbidden(c, err.Error())
		return
	}

	response.Success(c, files)
}
