package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	pb "thaily/proto/file"
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

// Excel model for storing Excel file info in MongoDB
type Excel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	File       string             `bson:"file" json:"file"`
	Title      string             `bson:"title" json:"title"`
	Option     string             `bson:"option" json:"option"`
	TableType  pb.TableType       `bson:"table_type" json:"table_type"`
	TableID    string             `bson:"table_id" json:"table_id"`
	Status     string             `bson:"status" json:"status"`
	Sum        int32              `bson:"sum" json:"sum"`
	Current    int32              `bson:"current" json:"current"`
	Messages   []any              `bson:"messages" json:"messages"`
	UploadType FileUploadType     `bson:"upload_type" json:"upload_type"`
	Percentage int32              `bson:"percentage" json:"percentage"`
	CreatedBy  string             `bson:"created_by" json:"created_by"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// BlobTokenClaims contains claims for temporary blob access token
type BlobTokenClaims struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// FileUploadResponse response for file upload
type FileUploadResponse struct {
	FileID       string `json:"file_id"`
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
	ObjectPath   string `json:"object_path"`
	UploadedBy   string `json:"uploaded_by"`
	UploadedRole string `json:"uploaded_role"`
}

// FileURLResponse response for file URL request
type FileURLResponse struct {
	FileID      string `json:"file_id"`
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ExpiresIn   string `json:"expires_in"`
}
