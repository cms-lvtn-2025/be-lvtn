package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"thaily/src/server/api/model"
)

// GetProtocol returns http or https based on TLS
func GetProtocol(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// ExtractObjectName extracts object name from file URL
func ExtractObjectName(fileURL string) string {
	parts := strings.Split(fileURL, "/")
	return strings.Join(parts, "/")
}

// ValidateFileType checks if file extension is valid for the upload type
func ValidateFileType(filename string, uploadType model.FileUploadType) error {
	ext := strings.ToLower(filepath.Ext(filename))

	switch uploadType {
	case model.UploadTypeGradeSuppervisor,
		model.UploadTypeGradeDefence,
		model.UploadTopicCouncilForDepartment,
		model.UploadCouncilForDepartment,
		model.UploadCouncilForAffair:
		// Accept xls, xlsx
		if ext != ".xls" && ext != ".xlsx" {
			return fmt.Errorf("list files must be .xls or .xlsx format")
		}
	case model.UploadTypeFinal, model.UploadTypeMidterm:
		// Accept pdf only
		if ext != ".pdf" {
			return fmt.Errorf("final files must be .pdf format")
		}
	}

	return nil
}

// GetContentType returns MIME type based on file extension
func GetContentType(filename string) string {
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

// GenerateObjectPath generates MinIO object path based on upload type and user info
func GenerateObjectPath(uploadType model.FileUploadType, userInfo *model.UserInfo, filename string, semester string) string {
	// Generate unique filename with timestamp and UUID
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	uniqueName := fmt.Sprintf("%s_%d_%s%s", baseName, time.Now().Unix(), uuid.New().String()[:8], ext)

	switch uploadType {
	case model.UploadTypeGradeSuppervisor:
		return fmt.Sprintf("grade_supervisor/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadTypeGradeDefence:
		return fmt.Sprintf("grade_defence/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadTopicCouncilForDepartment:
		return fmt.Sprintf("topic_for_department/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadCouncilForDepartment:
		return fmt.Sprintf("council_for_department/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadCouncilForAffair:
		return fmt.Sprintf("council_for_affair/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadStudentForAffair:
		return fmt.Sprintf("student_for_affair/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadTeacherForAffair:
		return fmt.Sprintf("teacher_for_affair/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadTypeMidterm:
		return fmt.Sprintf("midterm/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	case model.UploadTypeFinal:
		return fmt.Sprintf("final/%s/%s/%s", semester, userInfo.UserID, uniqueName)
	}

	return uniqueName
}

// GenerateBrowserFingerprint creates a unique fingerprint for the browser session
func GenerateBrowserFingerprint(c *gin.Context) string {
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
