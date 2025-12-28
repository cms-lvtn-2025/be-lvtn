package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	pb "thaily/proto/file"
	pbRole "thaily/proto/role"
	"thaily/proto/thesis"
	"thaily/src/server/api/helper"
	"thaily/src/server/api/model"
	"thaily/src/server/client"
	"thaily/src/server/config"
)

// FileService handles file business logic
type FileService struct {
	Config       *config.Config
	FileClient   *client.GRPCfile
	ThesisClient *client.GRPCthesis
	Mongodb      *client.MongoClient
	MinIO        *client.ServiceMinIo
	Redis        *client.RedisClient
	AuthService  *AuthService
}

// NewFileService creates a new FileService
func NewFileService(
	cfg *config.Config,
	fileClient *client.GRPCfile,
	thesisClient *client.GRPCthesis,
	mongodb *client.MongoClient,
	minIO *client.ServiceMinIo,
	redis *client.RedisClient,
	authService *AuthService,
) *FileService {
	return &FileService{
		Config:       cfg,
		FileClient:   fileClient,
		ThesisClient: thesisClient,
		Mongodb:      mongodb,
		MinIO:        minIO,
		Redis:        redis,
		AuthService:  authService,
	}
}

// UploadFileParams contains parameters for file upload
type UploadFileParams struct {
	File         multipart.File
	FileHeader   *multipart.FileHeader
	UploadType   model.FileUploadType
	UserInfo     *model.UserInfo
	Title        string
	TableType    string
	TableID      string
	EnrollmentID string
	SemesterCode string
}

// UploadFile handles file upload logic
func (s *FileService) UploadFile(ctx context.Context, c *gin.Context, params *UploadFileParams) (*model.FileUploadResponse, error) {
	// Validate file type
	if err := helper.ValidateFileType(params.FileHeader.Filename, params.UploadType); err != nil {
		return nil, fmt.Errorf("invalid file type: %w", err)
	}

	// Handle table type specific validation
	option, tableType, tableID, err := s.validateAndGetTableInfo(ctx, params)
	if err != nil {
		return nil, err
	}

	// Determine semester
	semester := params.UserInfo.Semester
	if params.UploadType == model.UploadCouncilForAffair ||
		params.UploadType == model.UploadStudentForAffair ||
		params.UploadType == model.UploadTeacherForAffair {
		semester = params.SemesterCode
	}

	// Generate object path
	objectPath := helper.GenerateObjectPath(params.UploadType, params.UserInfo, params.FileHeader.Filename, semester)
	contentType := helper.GetContentType(params.FileHeader.Filename)

	// Upload to MinIO
	fileURL, err := s.MinIO.UploadFile(ctx, objectPath, params.File, params.FileHeader.Size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Save file metadata
	var fileID string
	if option == "excel" {
		fileID, err = s.saveExcelToMongo(ctx, fileURL, params, tableType, option, tableID)
	} else {
		fileID, err = s.saveFileToGRPC(ctx, fileURL, params, tableType, option, tableID)
	}

	if err != nil {
		// Cleanup MinIO if save fails
		_ = s.MinIO.DeleteFile(ctx, objectPath)
		return nil, err
	}

	return &model.FileUploadResponse{
		FileID:       fileID,
		Filename:     params.FileHeader.Filename,
		Size:         params.FileHeader.Size,
		URL:          fileURL,
		ObjectPath:   objectPath,
		UploadedBy:   params.UserInfo.UserID,
		UploadedRole: params.UserInfo.Role,
	}, nil
}

func (s *FileService) validateAndGetTableInfo(ctx context.Context, params *UploadFileParams) (string, pb.TableType, string, error) {
	var option string
	var tableID string
	tableType := pb.TableType_TOPIC

	switch params.TableType {
	case "MIDTERM":
		tableType = pb.TableType_MIDTERM
		option = "midterm"
		tableID = params.TableID

		enrollment, err := s.ThesisClient.GetEnrollmentById(ctx, params.EnrollmentID)
		if err != nil {
			return "", 0, "", fmt.Errorf("failed to get enrollment: %w", err)
		}

		if enrollment.Enrollment.StudentCode != params.UserInfo.UserID {
			return "", 0, "", fmt.Errorf("you are not allowed to upload this file")
		}
		if *enrollment.Enrollment.MidtermCode != tableID {
			return "", 0, "", fmt.Errorf("you are not allowed to upload this file")
		}

		midterm, err := s.ThesisClient.GetMidtermById(ctx, *enrollment.Enrollment.MidtermCode)
		if err != nil {
			return "", 0, "", fmt.Errorf("failed to get midterm: %w", err)
		}
		if midterm.Midterm.Status != thesis.MidtermStatus_NOT_SUBMITTED {
			return "", 0, "", fmt.Errorf("you are not allowed to upload this file")
		}

	case "FINAL":
		tableType = pb.TableType_FINAL
		option = "final"
		tableID = params.TableID

		enrollment, err := s.ThesisClient.GetEnrollmentById(ctx, params.EnrollmentID)
		if err != nil {
			return "", 0, "", fmt.Errorf("failed to get enrollment: %w", err)
		}

		if enrollment.Enrollment.StudentCode != params.UserInfo.UserID {
			return "", 0, "", fmt.Errorf("you are not allowed to upload this file")
		}
		if *enrollment.Enrollment.FinalCode != tableID {
			return "", 0, "", fmt.Errorf("you are not allowed to upload this file")
		}

		final, err := s.ThesisClient.GetFinalById(ctx, *enrollment.Enrollment.FinalCode)
		if err != nil {
			return "", 0, "", fmt.Errorf("failed to get final: %w", err)
		}
		if final.Final.Status != thesis.FinalStatus_PENDING {
			return "", 0, "", fmt.Errorf("you are not allowed to upload this file")
		}

	case "ORDER":
		tableType = pb.TableType_ORDER
		option = "excel"
		tableID = "system"
	}

	return option, tableType, tableID, nil
}

func (s *FileService) saveExcelToMongo(ctx context.Context, fileURL string, params *UploadFileParams, tableType pb.TableType, option, tableID string) (string, error) {
	excelDoc := &model.Excel{
		File:       fileURL,
		Title:      params.Title,
		TableType:  tableType,
		Option:     option,
		UploadType: params.UploadType,
		Percentage: 0,
		Status:     "pending",
		Messages:   []any{},
		TableID:    tableID,
		CreatedBy:  params.UserInfo.UserID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := s.Mongodb.GetCollection("Excel").InsertOne(ctx, excelDoc)
	if err != nil {
		return "", fmt.Errorf("failed to save file to mongo db: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "excel", nil
}

func (s *FileService) saveFileToGRPC(ctx context.Context, fileURL string, params *UploadFileParams, tableType pb.TableType, option, tableID string) (string, error) {
	createResp, err := s.FileClient.CreateFile(ctx, &pb.CreateFileRequest{
		Title:     params.Title,
		File:      fileURL,
		Status:    pb.FileStatus_FILE_PENDING,
		Table:     tableType,
		Option:    option,
		TableId:   tableID,
		CreatedBy: params.UserInfo.UserID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save file to database: %w", err)
	}
	return createResp.File.Id, nil
}

// CheckRolePermission checks if user has required role
func (s *FileService) CheckRolePermission(ctx context.Context, userInfo *model.UserInfo, allowedRoles []string) (bool, error) {
	// Check direct role match
	for _, role := range allowedRoles {
		if userInfo.Role == role {
			return true, nil
		}
	}

	// For teachers, check extended roles
	if userInfo.Role == "teacher" {
		roles, err := s.AuthService.ExtractRole(ctx, userInfo.UserID)
		if err != nil {
			return false, err
		}
		for _, role := range roles {
			switch role {
			case pbRole.RoleType_TEACHER:
				if slices.Contains(allowedRoles, string("teacher")) {
					return true, nil
				}
			case pbRole.RoleType_DEPARTMENT_LECTURER:
				if slices.Contains(allowedRoles, string("department")) {
					return true, nil
				}
			case pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF:
				if slices.Contains(allowedRoles, string("affair")) {
					return true, nil
				}

			}

		}
	}

	return false, nil
}

// CanAccessFile checks if user has permission to access the file
func (s *FileService) CanAccessFile(fileResp *pb.File, userInfo *model.UserInfo) bool {
	// Owner can always access their own files
	if fileResp.CreatedBy == userInfo.UserID {
		return true
	}

	// Teacher can access files for review
	if userInfo.Role == "teacher" {
		return true
	}

	// Admin can access all files
	if userInfo.Role == "admin" {
		return true
	}

	return false
}

// GenerateBlobToken creates a temporary token for blob access
func (s *FileService) GenerateBlobToken(c *gin.Context, fileID string, userInfo *model.UserInfo) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	tokenID := uuid.New().String()

	claims := &model.BlobTokenClaims{
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
	tokenString, err := token.SignedString([]byte(s.Config.JWT.AccessSecret))
	if err != nil {
		return "", err
	}

	// Generate browser fingerprint
	fingerprint := helper.GenerateBrowserFingerprint(c)

	// Store token-fingerprint mapping in Redis
	redisKey := fmt.Sprintf("blob_token:%s", tokenID)
	err = s.Redis.Set(c.Request.Context(), redisKey, fingerprint, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to store token session: %w", err)
	}

	return tokenString, nil
}

// GetFileURL generates presigned URL or blob URL for file
func (s *FileService) GetFileURL(ctx context.Context, c *gin.Context, fileID string, userInfo *model.UserInfo, isExcel bool, urlType string) (*model.FileURLResponse, error) {
	var file, title string

	if isExcel {
		id, err := primitive.ObjectIDFromHex(fileID)
		if err != nil {
			return nil, fmt.Errorf("invalid file ID: %w", err)
		}

		var fileResp model.Excel
		result := s.Mongodb.GetCollection("Excel").FindOne(ctx, bson.M{
			"_id":        id,
			"created_by": userInfo.UserID,
		})
		if err := result.Decode(&fileResp); err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		file = fileResp.File
		title = fileResp.Title
	} else {
		fileResp, err := s.FileClient.GetFileById(ctx, fileID)
		if err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		if !s.CanAccessFile(fileResp.File, userInfo) {
			return nil, fmt.Errorf("you don't have permission to access this file")
		}
		file = fileResp.File.File
		title = fileResp.File.Title
	}

	var presignedURL, expiresIn string

	if urlType == "blob" {
		token, err := s.GenerateBlobToken(c, fileID, userInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}
		presignedURL = fmt.Sprintf("%s://%s/api/v1/files/blob?token=%s",
			helper.GetProtocol(c), c.Request.Host, token)
		expiresIn = "1 hour"
	} else {
		objectName := helper.ExtractObjectName(file)
		var err error
		presignedURL, err = s.MinIO.GetFileURL(ctx, objectName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate download URL: %w", err)
		}
		expiresIn = "5 minutes"
	}

	return &model.FileURLResponse{
		FileID:      fileID,
		DownloadURL: presignedURL,
		Filename:    title,
		ExpiresIn:   expiresIn,
	}, nil
}

// DeleteFile deletes a file from storage and database
func (s *FileService) DeleteFile(ctx context.Context, fileID string, userInfo *model.UserInfo, isExcel bool) error {
	var file string

	if isExcel {
		id, err := primitive.ObjectIDFromHex(fileID)
		if err != nil {
			return fmt.Errorf("invalid file ID: %w", err)
		}

		var fileResp model.Excel
		result := s.Mongodb.GetCollection("Excel").FindOne(ctx, bson.M{
			"_id":        id,
			"created_by": userInfo.UserID,
		})
		if err := result.Decode(&fileResp); err != nil {
			return fmt.Errorf("file not found: %w", err)
		}
		file = fileResp.File
	} else {
		fileResp, err := s.FileClient.GetFileById(ctx, fileID)
		if err != nil {
			return fmt.Errorf("file not found: %w", err)
		}
		if !s.CanAccessFile(fileResp.File, userInfo) {
			return fmt.Errorf("you don't have permission to access this file")
		}
		file = fileResp.File.File
	}

	// Delete from MinIO
	parts := strings.Split(file, "/")
	objectName := strings.Join(parts[0:], "/")
	_ = s.MinIO.DeleteFile(ctx, objectName)

	// Delete from database
	if isExcel {
		id, _ := primitive.ObjectIDFromHex(fileID)
		_, err := s.Mongodb.GetCollection("Excel").DeleteOne(ctx, bson.M{"_id": id})
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	} else {
		_, err := s.FileClient.DeleteFile(ctx, fileID)
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	return nil
}

// ListExcelFiles lists excel files based on role
func (s *FileService) ListExcelFiles(ctx context.Context, userInfo *model.UserInfo, roleHeader pbRole.RoleType) ([]model.Excel, error) {
	roles, err := s.AuthService.ExtractRole(ctx, userInfo.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract role: %w", err)
	}

	for _, role := range roles {
		if role == roleHeader {
			var uploadTypes []string
			switch roleHeader {
			case pbRole.RoleType_TEACHER:
				uploadTypes = []string{
					string(model.UploadTypeGradeSuppervisor),
					string(model.UploadTypeGradeDefence),
					string(model.UploadTopicCouncilForDepartment),
					string(model.UploadCouncilForDepartment),
					string(model.UploadCouncilForAffair),
				}
			case pbRole.RoleType_DEPARTMENT_LECTURER:
				uploadTypes = []string{
					string(model.UploadTopicCouncilForDepartment),
					string(model.UploadCouncilForDepartment),
					string(model.UploadCouncilForAffair),
				}
			case pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF:
				uploadTypes = []string{
					string(model.UploadCouncilForAffair),
					string(model.UploadStudentForAffair),
					string(model.UploadTeacherForAffair),
				}
			}

			cursor, err := s.Mongodb.GetCollection("Excel").Find(ctx, bson.M{
				"created_by": userInfo.UserID,
				"upload_type": bson.M{
					"$in": uploadTypes,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get excel: %w", err)
			}
			defer cursor.Close(ctx)

			var excelFiles []model.Excel
			if err := cursor.All(ctx, &excelFiles); err != nil {
				return nil, fmt.Errorf("failed to decode excel files: %w", err)
			}

			return excelFiles, nil
		}
	}

	return nil, fmt.Errorf("you are not allowed to list files")
}
