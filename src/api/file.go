package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"thaily/src/graph/helper"
	"thaily/src/pkg/response"
	"time"

	pb "thaily/proto/file"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// FileUploadType represents different upload destinations
type FileUploadType string

const (
	UploadTypeTemplate    FileUploadType = "template"     // tmp_template/{semester}/{teacher_id}
	UploadTypeListStudent FileUploadType = "list_student" // tmp_list_student/{semester}/{teacher_id}
	UploadTypeListTeacher FileUploadType = "list_teacher" // tmp_list_teacher/{semester}/{teacher_id}
	UploadTypeFinal       FileUploadType = "final"        // final/{semester}/{student_id}
)

// UserInfo contains extracted user information from JWT claims
type UserInfo struct {
	Role     string
	Semester string
	UserID   string
	IDs      []string
}

// BlobTokenClaims contains claims for temporary blob access token
type BlobTokenClaims struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// extractUserInfo extracts user information from context
func (h *APIHandler) extractUserInfo(c *gin.Context) (*UserInfo, error) {
	// Use c.Get() instead of c.Value() for gin.Context
	claimsValue, exists := c.Get(helper.Auth)
	if !exists {
		return nil, fmt.Errorf("not authorized - claims not found")
	}

	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("not authorized - invalid claims type")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("role not found in claims")
	}

	// Get semester from context (might be empty)
	semesterValue, _ := c.Get("semester")
	semester, _ := semesterValue.(string)

	// Parse IDs from claims
	idsStr, ok := claims["ids"].(string)
	if !ok {
		return nil, fmt.Errorf("ids not found in claims")
	}

	idsArr := strings.Split(idsStr, ",")
	myID := ""

	if semester == "" {
		// If no semester specified, use first ID
		if len(idsArr) > 0 {
			parts := strings.Split(idsArr[0], "-")
			if len(parts) > 1 {
				myID = parts[1]
				semester = parts[0]
			}
		}
	} else {
		// Find ID matching the semester
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+"-") {
				parts := strings.Split(id, "-")
				if len(parts) > 1 {
					myID = parts[1]
					break
				}
			}
		}
	}

	if myID == "" {
		return nil, fmt.Errorf("no user ID found for semester %s", semester)
	}

	return &UserInfo{
		Role:     role,
		Semester: semester,
		UserID:   myID,
		IDs:      idsArr,
	}, nil
}

// generateBrowserFingerprint creates a unique fingerprint for the browser session
func generateBrowserFingerprint(c *gin.Context) string {
	// Combine multiple factors to create unique fingerprint
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	xRealIP := c.GetHeader("X-Real-IP")

	// Create fingerprint string
	fingerprintData := fmt.Sprintf("%s|%s|%s|%s", userAgent, clientIP, xForwardedFor, xRealIP)

	// Hash it for security
	hash := sha256.Sum256([]byte(fingerprintData))
	return hex.EncodeToString(hash[:])
}

// generateBlobToken creates a temporary token for blob access bound to browser session
func (h *APIHandler) generateBlobToken(c *gin.Context, fileID string, userInfo *UserInfo) (string, error) {
	// Token valid for 1 hour
	expirationTime := time.Now().Add(1 * time.Hour)
	tokenID := uuid.New().String()

	claims := &BlobTokenClaims{
		FileID: fileID,
		UserID: userInfo.UserID,
		Role:   userInfo.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.Config.JWT.AccessSecret))
	if err != nil {
		return "", err
	}

	// Generate browser fingerprint
	fingerprint := generateBrowserFingerprint(c)

	// Store token-fingerprint mapping in Redis
	redisKey := fmt.Sprintf("blob_token:%s", tokenID)
	err = h.Redis.Set(c.Request.Context(), redisKey, fingerprint, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to store token session: %w", err)
	}

	return tokenString, nil
}

// validateBlobToken validates token and checks browser fingerprint
func (h *APIHandler) validateBlobToken(c *gin.Context, tokenString string) (*BlobTokenClaims, error) {
	claims := &BlobTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.Config.JWT.AccessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Get stored fingerprint from Redis
	redisKey := fmt.Sprintf("blob_token:%s", claims.ID)
	storedFingerprint, err := h.Redis.Get(c.Request.Context(), redisKey)
	if err != nil {
		return nil, fmt.Errorf("token session expired or invalid")
	}

	// Generate fingerprint from current request
	currentFingerprint := generateBrowserFingerprint(c)

	// Compare fingerprints - must match
	if storedFingerprint != currentFingerprint {
		return nil, fmt.Errorf("token cannot be used from different browser/session")
	}

	return claims, nil
}

// canAccessFile checks if user has permission to access the file
func (h *APIHandler) canAccessFile(fileResp *pb.File, userInfo *UserInfo) bool {
	// Owner can always access their own files
	if fileResp.CreatedBy == userInfo.UserID {
		return true
	}

	// Teacher can access files in their assigned classes/topics
	if userInfo.Role == "teacher" {
		// Teachers can view student files for review
		// You can add more specific logic here based on table_id, semester, etc.
		return true
	}

	// Admin can access all files (if you have admin role)
	if userInfo.Role == "admin" {
		return true
	}

	// Students can only access their own files
	return false
}

// validateFileType checks if file extension is valid for the upload type
func validateFileType(filename string, uploadType FileUploadType) error {
	ext := strings.ToLower(filepath.Ext(filename))

	switch uploadType {
	case UploadTypeTemplate:
		// Accept doc, docx
		if ext != ".doc" && ext != ".docx" {
			return fmt.Errorf("template files must be .doc or .docx format")
		}
	case UploadTypeListStudent, UploadTypeListTeacher:
		// Accept xls, xlsx
		if ext != ".xls" && ext != ".xlsx" {
			return fmt.Errorf("list files must be .xls or .xlsx format")
		}
	case UploadTypeFinal:
		// Accept pdf only
		if ext != ".pdf" {
			return fmt.Errorf("final files must be .pdf format")
		}
	}

	return nil
}

// getContentType returns MIME type based on file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

// generateObjectPath generates MinIO object path based on upload type and user info
func generateObjectPath(uploadType FileUploadType, userInfo *UserInfo, filename string) string {
	// Generate unique filename with timestamp and UUID
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	uniqueName := fmt.Sprintf("%s_%d_%s%s", baseName, time.Now().Unix(), uuid.New().String()[:8], ext)

	switch uploadType {
	case UploadTypeTemplate:
		return fmt.Sprintf("tmp_template/%s/%s/%s", userInfo.Semester, userInfo.UserID, uniqueName)
	case UploadTypeListStudent:
		return fmt.Sprintf("tmp_list_student/%s/%s/%s", userInfo.Semester, userInfo.UserID, uniqueName)
	case UploadTypeListTeacher:
		return fmt.Sprintf("tmp_list_teacher/%s/%s/%s", userInfo.Semester, userInfo.UserID, uniqueName)
	case UploadTypeFinal:
		return fmt.Sprintf("final/%s/%s/%s", userInfo.Semester, userInfo.UserID, uniqueName)
	}

	return uniqueName
}

// uploadFileHandler handles file upload with validation
func (h *APIHandler) uploadFileHandler(c *gin.Context, uploadType FileUploadType, allowedRoles []string) {
	// Extract user info
	userInfo, err := h.extractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Check role permission
	roleAllowed := false
	for _, role := range allowedRoles {
		if userInfo.Role == role {
			roleAllowed = true
			break
		}
	}
	if !roleAllowed {
		response.Forbidden(c, fmt.Sprintf("role %s is not allowed to upload this type of file", userInfo.Role))
		return
	}

	// Check semester for certain upload types
	if uploadType != UploadTypeFinal && userInfo.Semester == "" {
		response.BadRequest(c, "semester is required")
		return
	}

	// Get file from request
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "no file uploaded")
		return
	}

	// Validate file type
	if err := validateFileType(fileHeader.Filename, uploadType); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c, "failed to open uploaded file")
		return
	}
	defer file.Close()

	// Generate object path
	objectPath := generateObjectPath(uploadType, userInfo, fileHeader.Filename)
	contentType := getContentType(fileHeader.Filename)

	// Upload to MinIO
	fileURL, err := h.MimIo.UploadFile(c.Request.Context(), objectPath, file, fileHeader.Size, contentType)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("failed to upload file: %v", err))
		return
	}

	// Get optional fields
	title := c.PostForm("title")
	if title == "" {
		title = fileHeader.Filename
	}

	tableType := pb.TableType_TOPIC
	tableTypeStr := c.PostForm("table_type")
	switch tableTypeStr {
	case "TOPIC":
		tableType = pb.TableType_TOPIC
	case "MIDTERM":
		tableType = pb.TableType_MIDTERM
	case "FINAL":
		tableType = pb.TableType_FINAL
	case "ORDER":
		tableType = pb.TableType_ORDER
	}

	// Get option with default value based on upload type
	option := c.PostForm("option")
	if option == "" {
		switch uploadType {
		case UploadTypeTemplate:
			option = "template"
		case UploadTypeListStudent:
			option = "student_list"
		case UploadTypeListTeacher:
			option = "teacher_list"
		case UploadTypeFinal:
			option = "final_document"
		default:
			option = "general"
		}
	}

	tableID := c.PostForm("table_id")
	if tableID == "" {
		tableID = "system"
	}

	// Save file metadata to database via gRPC
	createResp, err := h.FileClient.CreateFile(c.Request.Context(), &pb.CreateFileRequest{
		Title:     title,
		File:      fileURL,
		Status:    pb.FileStatus_REJECTED,
		Table:     tableType,
		Option:    option,
		TableId:   tableID,
		CreatedBy: userInfo.UserID,
	})

	if err != nil {
		// If database save fails, try to delete from MinIO
		_ = h.MimIo.DeleteFile(c.Request.Context(), objectPath)
		response.InternalError(c, fmt.Sprintf("failed to save file metadata: %v", err))
		return
	}

	// Invalidate file cache
	_ = h.FileClient.InvalidateAllFileCache(c.Request.Context())

	response.SuccessWithMessage(c, "File uploaded successfully", gin.H{
		"file_id":       createResp.File.Id,
		"filename":      fileHeader.Filename,
		"size":          fileHeader.Size,
		"url":           fileURL,
		"object_path":   objectPath,
		"uploaded_by":   userInfo.UserID,
		"uploaded_role": userInfo.Role,
	})
}

// UploadTemplateFile handles template document upload (Word files)
// POST /api/files/upload/template
func (h *APIHandler) UploadTemplateFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeTemplate, []string{"teacher"})
}

// UploadListStudentFile handles student list upload (Excel files)
// POST /api/files/upload/list-student
func (h *APIHandler) UploadListStudentFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeListStudent, []string{"teacher"})
}

// UploadListTeacherFile handles teacher list upload (Excel files)
// POST /api/files/upload/list-teacher
func (h *APIHandler) UploadListTeacherFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeListTeacher, []string{"teacher"})
}

// UploadFinalFile handles final document upload (PDF files)
// POST /api/files/upload/final
func (h *APIHandler) UploadFinalFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeFinal, []string{"student"})
}

// GetFile retrieves file information by ID
// GET /api/files/:id
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

	// Get file from database
	fileResp, err := h.FileClient.GetFileById(c.Request.Context(), fileID)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("File not found: %v", err))
		return
	}

	response.Success(c, gin.H{
		"file": fileResp.File,
	})
}

// GetFileURL generates a presigned URL for file download
// GET /api/files/:id/url
func (h *APIHandler) GetFileURL(c *gin.Context) {
	if h.FileClient == nil || h.MimIo == nil {
		response.InternalError(c, "File service not available")
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	// Get file metadata
	fileResp, err := h.FileClient.GetFileById(c.Request.Context(), fileID)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("File not found: %v", err))
		return
	}

	// Extract object path from file URL
	fileURL := fileResp.File.File
	parts := strings.Split(fileURL, "/")
	if len(parts) < 5 {
		response.InternalError(c, "Invalid file URL format")
		return
	}

	// URL format: http://host:port/bucket/path/to/file
	// parts: [http:, , host:port, bucket, path, to, file]
	// We need everything after bucket (index 4 onwards)
	objectName := strings.Join(parts[4:], "/")

	// Generate presigned URL
	presignedURL, err := h.MimIo.GetFileURL(c.Request.Context(), objectName)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed to generate download URL: %v", err))
		return
	}

	response.Success(c, gin.H{
		"file_id":      fileID,
		"download_url": presignedURL,
		"filename":     fileResp.File.Title,
		"expires_in":   "7 days",
	})
}

// DeleteFile deletes a file
// DELETE /api/files/:id
func (h *APIHandler) DeleteFile(c *gin.Context) {
	if h.FileClient == nil || h.MimIo == nil {
		response.InternalError(c, "File service not available")
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	// Extract user info for authorization
	userInfo, err := h.extractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Get file metadata
	fileResp, err := h.FileClient.GetFileById(c.Request.Context(), fileID)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("File not found: %v", err))
		return
	}

	// Check if user is the owner
	if fileResp.File.CreatedBy != userInfo.UserID {
		response.Forbidden(c, "You can only delete your own files")
		return
	}

	// Extract object path from file URL
	fileURL := fileResp.File.File
	parts := strings.Split(fileURL, "/")
	if len(parts) >= 5 {
		// URL format: http://host:port/bucket/path/to/file
		// parts: [http:, , host:port, bucket, path, to, file]
		// We need everything after bucket (index 4 onwards)
		objectName := strings.Join(parts[4:], "/")
		// Delete from MinIO (ignore error if file doesn't exist)
		_ = h.MimIo.DeleteFile(c.Request.Context(), objectName)
	}

	// Delete from database (implement DeleteFile in client if not exists)
	// For now, we can update status to deleted or implement DeleteFile

	// Invalidate cache
	_ = h.FileClient.InvalidateFileCache(c.Request.Context(), fileID)

	response.SuccessWithMessage(c, "File deleted successfully", gin.H{
		"file_id": fileID,
	})
}

// ListFiles lists files with filtering
// GET /api/files
func (h *APIHandler) ListFiles(c *gin.Context) {
	if h.FileClient == nil {
		response.InternalError(c, "File service not available")
		return
	}

	// Extract user info
	userInfo, err := h.extractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Build search request (you can add more filters from query params)
	// For now, return user's files only
	_ = userInfo // Use userInfo to filter files

	// TODO: Build proper search request with filters
	// files, err := h.FileClient.GetFileBySearch(c.Request.Context(), searchRequest)

	response.Success(c, gin.H{
		"message": "List files endpoint - implement search logic",
	})
}

// GetBlobURL generates a temporary blob URL with token
// GET /api/files/:id/blob-url
func (h *APIHandler) GetBlobURL(c *gin.Context) {
	if h.FileClient == nil {
		response.InternalError(c, "File service not available")
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	// Extract user info for authorization
	userInfo, err := h.extractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Get file metadata
	fileResp, err := h.FileClient.GetFileById(c.Request.Context(), fileID)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("File not found: %v", err))
		return
	}

	// Check if user can access this file
	if !h.canAccessFile(fileResp.File, userInfo) {
		response.Forbidden(c, "You don't have permission to access this file")
		return
	}

	// Generate blob token (bound to current browser session)
	token, err := h.generateBlobToken(c, fileID, userInfo)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed to generate token: %v", err))
		return
	}

	// Build blob URL
	blobURL := fmt.Sprintf("%s://%s/api/v1/files/blob?token=%s",
		func() string {
			if c.Request.TLS != nil {
				return "https"
			}
			return "http"
		}(),
		c.Request.Host,
		token,
	)

	response.Success(c, gin.H{
		"file_id":    fileID,
		"blob_url":   blobURL,
		"filename":   fileResp.File.Title,
		"expires_in": "1 hour",
	})
}

// GetFileBlob serves file content directly using token
// GET /api/files/blob?token=xxx
func (h *APIHandler) GetFileBlob(c *gin.Context) {
	if h.FileClient == nil || h.MimIo == nil {
		response.InternalError(c, "File service not available")
		return
	}

	// Get token from query parameter
	tokenString := c.Query("token")
	if tokenString == "" {
		response.Unauthorized(c, "Token required")
		return
	}

	// Validate token and check browser fingerprint
	claims, err := h.validateBlobToken(c, tokenString)
	if err != nil {
		response.Unauthorized(c, fmt.Sprintf("Invalid token: %v", err))
		return
	}

	// Get file metadata
	fileResp, err := h.FileClient.GetFileById(c.Request.Context(), claims.FileID)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("File not found: %v", err))
		return
	}

	// Verify token file ID matches
	if fileResp.File.Id != claims.FileID {
		response.Forbidden(c, "Token does not match file")
		return
	}

	// Extract object path from file URL
	fileURL := fileResp.File.File
	parts := strings.Split(fileURL, "/")
	if len(parts) < 5 {
		response.InternalError(c, "Invalid file URL format")
		return
	}

	// URL format: http://host:port/bucket/path/to/file
	// parts: [http:, , host:port, bucket, path, to, file]
	// We need everything after bucket (index 4 onwards)
	objectName := strings.Join(parts[4:], "/")

	// Get file blob from MinIO
	object, objectInfo, err := h.MimIo.GetFileBlob(c.Request.Context(), objectName)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed to retrieve file: %v", err))
		return
	}
	defer object.Close()

	// Set headers for file serving
	c.Header("Content-Type", objectInfo.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", objectInfo.Size))
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileResp.File.Title))
	c.Header("Cache-Control", "public, max-age=3600")

	// Stream file to response
	c.DataFromReader(200, objectInfo.Size, objectInfo.ContentType, object, nil)
}
