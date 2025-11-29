package api

import (
	"fmt"
	"slices"
	"strings"
	"thaily/src/server/response"
	"time"

	pb "thaily/proto/file"
	"thaily/proto/thesis"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileUploadType represents different upload destinations
type FileUploadType string

const (
	UploadTypeGradeSuppervisor      FileUploadType = "grade_supervisor"       // grade_supervisor/{semester}/{student_id}
	UploadTypeGradeDefence          FileUploadType = "grade_defence"          // grade_defence/{semester}/{student_id}
	UploadTopicCouncilForDepartment FileUploadType = "topic_for_department"   // topic_for_department/{semester}/{teacher_id}
	UploadCouncilForDepartment      FileUploadType = "council_for_department" // council_for_department/{semester}/{teacher_id}
	UploadCouncilForAffair          FileUploadType = "council_for_affair"     // council_for_affair/{semester}/{teacher_id}
	UploadStudentForAffair          FileUploadType = "student_for_affair"     // student_for_affair/{semester}/{student_id}
	UploadTeacherForAffair          FileUploadType = "teacher_for_affair"     // teacher_for_affair/{semester}/{teacher_id}
	UploadTypeMidterm               FileUploadType = "midterm"                // midterm/{semester}/{student_id}
	UploadTypeFinal                 FileUploadType = "final"                  // final/{semester}/{student_id}
)

// Excel model lưu thông tin file Excel trong MongoDB
type Excel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	File       string             `bson:"file" json:"file"`
	Title      string             `bson:"title" json:"title"`
	Option     string             `bson:"option" json:"option"`
	TableType  pb.TableType       `bson:"table_type" json:"table_type"`
	TableID    string             `bson:"table_id" json:"table_id"`
	Status     string             `bson:"status" json:"status"`
	Message    string             `bson:"message" json:"message"`
	UploadType FileUploadType     `bson:"upload_type" json:"upload_type"`
	Percentage int32              `bson:"percentage" json:"percentage"`
	CreatedBy  string             `bson:"created_by" json:"created_by"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

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
	fingerprint := GenerateBrowserFingerprint(c)

	// Store token-fingerprint mapping in Redis
	redisKey := fmt.Sprintf("blob_token:%s", tokenID)
	err = h.Redis.Set(c.Request.Context(), redisKey, fingerprint, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to store token session: %w", err)
	}

	return tokenString, nil
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
	if userInfo.Role == "teacher" && !roleAllowed {
		roles, err := h.extractRole(c, userInfo.UserID)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("failed to extract role: %v", err))
			return
		}
		for _, role := range roles {
			if slices.Contains(allowedRoles, string(role)) {
				roleAllowed = true
				break
			}
		}
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
	if err := ValidateFileType(fileHeader.Filename, uploadType); err != nil {
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

	// Get optional fields
	title := c.PostForm("title")
	if title == "" {
		title = fileHeader.Filename
	}
	var option string
	var tableID string
	tableType := pb.TableType_TOPIC
	tableTypeStr := c.PostForm("table_type")
	switch tableTypeStr {
	case "MIDTERM":
		tableType = pb.TableType_MIDTERM
		option = "midterm"
		tableID = c.PostForm("table_id")
		enrollmentID := c.PostForm("enrollment_id")
		// check student_code == user_id
		enrollment, err := h.ThesisClient.GetEnrollmentById(c.Request.Context(), enrollmentID)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("failed to get enrollment: %v", err))
			return
		}

		if enrollment.Enrollment.StudentCode != userInfo.UserID {
			response.Forbidden(c, "you are not allowed 1 to upload this file")
			return
		}
		if *enrollment.Enrollment.MidtermCode != tableID {
			response.Forbidden(c, "you are not allowed 2 to upload this file")
			return
		}
		// getMidterm check status is not submitted
		midterm, err := h.ThesisClient.GetMidtermById(c.Request.Context(), *enrollment.Enrollment.MidtermCode)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("failed to get midterm: %v", err))
			return
		}
		if midterm.Midterm.Status != thesis.MidtermStatus_NOT_SUBMITTED {
			response.Forbidden(c, "you are not allowed to upload this file")
			return
		}
	case "FINAL":
		tableType = pb.TableType_FINAL
		option = "final"
		tableID = c.PostForm("table_id")
		enrollmentID := c.PostForm("enrollment_id")
		enrollment, err := h.ThesisClient.GetEnrollmentById(c.Request.Context(), enrollmentID)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("failed to get enrollment: %v", err))
			return
		}
		if enrollment.Enrollment.StudentCode != userInfo.UserID {
			response.Forbidden(c, "you are not allowed to upload this file")
			return
		}
		if *enrollment.Enrollment.FinalCode != tableID {
			response.Forbidden(c, "you are not allowed to upload this file")
			return
		}
		// check status final
		final, err := h.ThesisClient.GetFinalById(c.Request.Context(), *enrollment.Enrollment.FinalCode)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("failed to get final: %v", err))
			return
		}
		if final.Final.Status != thesis.FinalStatus_PENDING {
			response.Forbidden(c, "you are not allowed to upload this file")
			return
		}
	case "ORDER":
		tableType = pb.TableType_ORDER
		option = "excel"
		tableID = "system"
	}

	objectPath := GenerateObjectPath(uploadType, userInfo, fileHeader.Filename)
	contentType := GetContentType(fileHeader.Filename)

	// Upload to MinIO
	fileURL, err := h.MimIo.UploadFile(c.Request.Context(), objectPath, file, fileHeader.Size, contentType)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("failed to upload file: %v", err))
		return
	}

	// Get option with default value based on upload type
	// excel no save on database but save mongo db

	// Save file metadata to database via gRPC
	var fileID string
	if option == "excel" {
		// save to mongo db
		excelDoc := &Excel{
			File:       fileURL,
			Title:      title,
			TableType:  tableType,
			Option:     option,
			UploadType: uploadType,
			Percentage: 0,
			Status:     "pending",
			Message:    "",
			TableID:    tableID,
			CreatedBy:  userInfo.UserID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		result, err := h.Mongodb.GetCollection("Excel").InsertOne(c.Request.Context(), excelDoc)
		if err != nil {
			// If database save fails, try to delete from MinIO
			_ = h.MimIo.DeleteFile(c.Request.Context(), objectPath)
			response.InternalError(c, fmt.Sprintf("failed to save file to mongo db: %v", err))
			return
		}
		// For excel files, use MongoDB ObjectID as file ID
		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
			fileID = oid.Hex()
		} else {
			fileID = "excel"
		}
	} else {
		createResp, err := h.FileClient.CreateFile(c.Request.Context(), &pb.CreateFileRequest{
			Title:     title,
			File:      fileURL,
			Status:    pb.FileStatus_FILE_PENDING,
			Table:     tableType,
			Option:    option,
			TableId:   tableID,
			CreatedBy: userInfo.UserID,
		})
		if err != nil {
			// If database save fails, try to delete from MinIO
			_ = h.MimIo.DeleteFile(c.Request.Context(), objectPath)
			response.InternalError(c, fmt.Sprintf("failed to save file to database: %v", err))
			return
		}
		fileID = createResp.File.Id
	}

	// Invalidate file cache
	//_ = h.FileClient.InvalidateAllFileCache(c.Request.Context())

	response.SuccessWithMessage(c, "File uploaded successfully", gin.H{
		"file_id":       fileID,
		"filename":      fileHeader.Filename,
		"size":          fileHeader.Size,
		"url":           fileURL,
		"object_path":   objectPath,
		"uploaded_by":   userInfo.UserID,
		"uploaded_role": userInfo.Role,
	})
}

// UploadGradeSuppervisorFile handles grade supervisor upload (Excel files)
// POST /api/files/upload/grade-supervisor
func (h *APIHandler) UploadGradeSuppervisorFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeGradeSuppervisor, []string{"teacher"})
}

// UploadGradeDefenceFile handles grade defence upload (Excel files)
// POST /api/files/upload/grade-defence
func (h *APIHandler) UploadGradeDefenceFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeGradeDefence, []string{"teacher"})
}

// UploadTopicCouncilForDepartmentFile handles topic council for department upload (Excel files)
// POST /api/files/upload/topic-council-for-department
func (h *APIHandler) UploadTopicCouncilForDepartmentFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTopicCouncilForDepartment, []string{"department"})
}

// UploadCouncilForDepartmentFile handles council for department upload (Excel files)
// POST /api/files/upload/council-for-department
func (h *APIHandler) UploadCouncilForDepartmentFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadCouncilForDepartment, []string{"department"})
}

// UploadCouncilForAffairFile handles council for affair upload (Excel files)
// POST /api/files/upload/council-for-affair
func (h *APIHandler) UploadCouncilForAffairFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadCouncilForAffair, []string{"affair"})
}

// UploadUserForAffairFile handles user for affair upload (Excel files)
// POST /api/files/upload/user-for-affair
func (h *APIHandler) UploadUserForAffairFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadStudentForAffair, []string{"affair"})
}

// UploadTeacherForAffairFile handles teacher for affair upload (Excel files)
// POST /api/files/upload/teacher-for-affair
func (h *APIHandler) UploadTeacherForAffairFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTeacherForAffair, []string{"affair"})
}

// UploadMidtermFile handles midterm upload (PDF files)
// POST /api/files/upload/midterm
func (h *APIHandler) UploadMidtermFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeMidterm, []string{"student"})
}

// UploadFinalFile handles final upload (PDF files)
// POST /api/files/upload/final
func (h *APIHandler) UploadFinalFile(c *gin.Context) {
	h.uploadFileHandler(c, UploadTypeFinal, []string{"student"})
}

// GetFileURL generates a presigned URL for file download
// GET /api/files/:id/url
func (h *APIHandler) GetFileURL(c *gin.Context) {
	if h.FileClient == nil {
		response.InternalError(c, "File service not available")
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		response.BadRequest(c, "File ID required")
		return
	}

	// ← THÊM 5 DÒNG NÀY (TỪ GetBlobURL)
	userInfo, err := h.extractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	fileResp, err := h.FileClient.GetFileById(c.Request.Context(), fileID)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("File not found: %v", err))
		return
	}

	if !h.canAccessFile(fileResp.File, userInfo) {
		response.Forbidden(c, "You don't have permission to access this file")
		return
	}
	// ← END 5 DÒNG

	// ← THÊM 1 DÒNG: PHÂN BIỆT TYPE
	urlType := c.Query("type") // "blob" hoặc "" (default)

	var presignedURL, expiresIn string

	if urlType == "blob" {
		// ← THÊM 4 DÒNG: BLOB LOGIC
		token, err := h.generateBlobToken(c, fileID, userInfo)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("Failed to generate token: %v", err))
			return
		}
		presignedURL = fmt.Sprintf("%s://%s/api/v1/files/blob?token=%s",
			GetProtocol(c), c.Request.Host, token)
		expiresIn = "1 hour"
	} else {
		// CODE CŨ: PRESIGNED URL
		objectName := ExtractObjectName(fileResp.File.File)
		presignedURL, err = h.MimIo.GetFileURL(c.Request.Context(), objectName)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("Failed to generate download URL: %v", err))
			return
		}
		expiresIn = "5 minutes"
	}

	response.Success(c, gin.H{
		"file_id":      fileID,
		"download_url": presignedURL, // ← THỐNG NHẤT TÊN FIELD
		"filename":     fileResp.File.Title,
		"expires_in":   expiresIn,
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

	// URL format: /bucket/path/to/file
	// parts: [http:, , host:port, bucket, path, to, file]
	// We need everything after bucket (index 4 onwards)
	objectName := strings.Join(parts[0:], "/")
	// Delete from MinIO (ignore error if file doesn't exist)
	_ = h.MimIo.DeleteFile(c.Request.Context(), objectName)
	_, err = h.FileClient.DeleteFile(c.Request.Context(), fileID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed to delete file: %v", err))
		return
	}

	// Delete from database (implement DeleteFile in client if not exists)
	// For now, we can update status to deleted or implement DeleteFile

	// Invalidate cache
	//_ = h.FileClient.InvalidateFileCache(c.Request.Context(), fileID)

	response.SuccessWithMessage(c, "File deleted successfully", gin.H{
		"file_id": fileID,
	})
}

// ListFiles lists files with filtering
// GET /api/files from mongodb fllow role and semester with x-role
func (h *APIHandler) ListFilesExcel(c *gin.Context) {
	userInfo, err := h.extractUserInfo(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roleHeader := c.GetHeader("X-Role")
	roles, err := h.extractRole(c, userInfo.UserID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("failed to extract role: %v", err))
		return
	}
	for _, role := range roles {
		if role == roleHeader {
			// get files from mongodb follow role and semester
			var uploadTypes []FileUploadType
			switch roleHeader {
			case "teacher":
				uploadTypes = []FileUploadType{UploadTypeGradeSuppervisor, UploadTypeGradeDefence, UploadTopicCouncilForDepartment, UploadCouncilForDepartment, UploadCouncilForAffair}
			case "department":
				uploadTypes = []FileUploadType{UploadTopicCouncilForDepartment, UploadCouncilForDepartment, UploadCouncilForAffair}
			case "affair":
				uploadTypes = []FileUploadType{UploadCouncilForAffair, UploadStudentForAffair, UploadTeacherForAffair}
			}
			excel, err := h.Mongodb.GetCollection("Excel").Find(c.Request.Context(), bson.M{
				"created_by": userInfo.UserID,
				"upload_type": bson.M{
					"$in": uploadTypes,
				},
			})
			if err != nil {
				response.InternalError(c, fmt.Sprintf("failed to get excel: %v", err))
				return
			}
			response.Success(c, excel)
		}
	}
	response.Forbidden(c, "You are not allowed to list files")

}
