# Minio Integration Guide

## Tổng quan

Minio là object storage tương thích với S3 API, được sử dụng để lưu trữ files trong hệ thống.

## Kiến trúc

```
Client → Backend Gateway → File Service → Minio
                                ↓
                           Database (metadata)
```

## Setup Minio

### 1. Cài đặt với Docker

```bash
# Chạy Minio
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  -v minio-data:/data \
  minio/minio server /data --console-address ":9001"
```

### 2. Truy cập Minio Console

URL: http://localhost:9001
Username: `minioadmin`
Password: `minioadmin`

### 3. Tạo buckets

```bash
# Sử dụng mc (Minio Client)
mc alias set myminio http://localhost:9000 minioadmin minioadmin

# Tạo buckets
mc mb myminio/avatars
mc mb myminio/documents
mc mb myminio/thesis
mc mb myminio/academic

# Set public read cho avatars
mc anonymous set download myminio/avatars
```

## Integration Code

### 1. Cài đặt Minio SDK

```bash
go get github.com/minio/minio-go/v7
```

### 2. Tạo Minio Client Package

**File**: `src/pkg/storage/minio.go`

```go
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client *minio.Client
	config MinioConfig
}

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	Region     string
}

func NewMinioClient(cfg MinioConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinioClient{
		client: client,
		config: cfg,
	}, nil
}

// UploadFile uploads a file to Minio
func (m *MinioClient) UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// Ensure bucket exists
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
			Region: m.config.Region,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Upload file
	info, err := m.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, bucket, objectName)

	return url, nil
}

// GetFile gets file info
func (m *MinioClient) GetFile(ctx context.Context, bucket, objectName string) (*minio.ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &info, nil
}

// GetPresignedURL generates a pre-signed URL for temporary access
func (m *MinioClient) GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}

	return url.String(), nil
}

// DeleteFile deletes a file from Minio
func (m *MinioClient) DeleteFile(ctx context.Context, bucket, objectName string) error {
	err := m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListFiles lists files in a bucket
func (m *MinioClient) ListFiles(ctx context.Context, bucket, prefix string) ([]string, error) {
	var files []string

	for object := range m.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}
```

### 3. Thêm Minio vào Container

**File**: `src/pkg/container/container.go`

```go
package container

import (
	"thaily/src/config"
	"thaily/src/pkg/storage"
	"thaily/src/server/client"
)

type Container struct {
	Config   *config.Config
	Clients  *Clients
	Storage  *storage.MinioClient  // Thêm Minio
}

func New(cfg *config.Config) (*Container, error) {
	// ... existing code

	// Initialize Minio
	minioClient, err := storage.NewMinioClient(storage.MinioConfig{
		Endpoint:  cfg.Minio.Endpoint,
		AccessKey: cfg.Minio.AccessKey,
		SecretKey: cfg.Minio.SecretKey,
		UseSSL:    cfg.Minio.UseSSL,
		Region:    cfg.Minio.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio: %w", err)
	}

	return &Container{
		Config:  cfg,
		Clients: clients,
		Storage: minioClient,  // Add to container
	}, nil
}
```

### 4. Cập nhật Config

**File**: `src/config/config.go`

```go
type Config struct {
	Server   ServerConfig
	Services ServiceConfig
	Google   GoogleOAuthConfig
	Minio    MinioConfig  // Thêm Minio config
}

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	Region     string
	Buckets    BucketConfig
}

type BucketConfig struct {
	Avatars   string
	Documents string
	Thesis    string
	Academic  string
}

func Load() (*Config, error) {
	// ...

	cfg := &Config{
		// ... existing configs

		Minio: MinioConfig{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			UseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
			Region:    getEnv("MINIO_REGION", "us-east-1"),
			Buckets: BucketConfig{
				Avatars:   getEnv("MINIO_BUCKET_AVATARS", "avatars"),
				Documents: getEnv("MINIO_BUCKET_DOCUMENTS", "documents"),
				Thesis:    getEnv("MINIO_BUCKET_THESIS", "thesis"),
				Academic:  getEnv("MINIO_BUCKET_ACADEMIC", "academic"),
			},
		},
	}

	return cfg, nil
}
```

### 5. Sử dụng trong API Handler

**File**: `src/api/file.go`

```go
package api

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"thaily/src/pkg/response"
	"thaily/src/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	FileClient  *client.GRPCfile
	MinioClient *storage.MinioClient
	Config      *config.Config
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded")
		return
	}

	// Get bucket from form (default: documents)
	bucket := c.DefaultPostForm("bucket", "documents")

	// Validate bucket
	validBuckets := map[string]bool{
		h.Config.Minio.Buckets.Avatars:   true,
		h.Config.Minio.Buckets.Documents: true,
		h.Config.Minio.Buckets.Thesis:    true,
		h.Config.Minio.Buckets.Academic:  true,
	}

	if !validBuckets[bucket] {
		response.BadRequest(c, "Invalid bucket")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Open file
	src, err := file.Open()
	if err != nil {
		response.InternalError(c, "Failed to open file")
		return
	}
	defer src.Close()

	// Upload to Minio
	ctx := context.Background()
	url, err := h.MinioClient.UploadFile(
		ctx,
		bucket,
		objectName,
		src,
		file.Size,
		file.Header.Get("Content-Type"),
	)
	if err != nil {
		response.InternalError(c, "Failed to upload file")
		return
	}

	// TODO: Save metadata to database via FileClient
	// fileMetadata := &pb.FileMetadata{
	// 	Name:     file.Filename,
	// 	Size:     file.Size,
	// 	Url:      url,
	// 	Bucket:   bucket,
	// 	ObjectName: objectName,
	// }
	// h.FileClient.SaveMetadata(ctx, fileMetadata)

	response.SuccessWithMessage(c, "File uploaded successfully", gin.H{
		"filename":    file.Filename,
		"size":        file.Size,
		"url":         url,
		"object_name": objectName,
		"bucket":      bucket,
	})
}

func (h *FileHandler) GetFileURL(c *gin.Context) {
	bucket := c.Query("bucket")
	objectName := c.Query("object")

	if bucket == "" || objectName == "" {
		response.BadRequest(c, "bucket and object are required")
		return
	}

	// Generate pre-signed URL (valid for 1 hour)
	ctx := context.Background()
	url, err := h.MinioClient.GetPresignedURL(ctx, bucket, objectName, 1*time.Hour)
	if err != nil {
		response.InternalError(c, "Failed to generate URL")
		return
	}

	response.Success(c, gin.H{
		"url":     url,
		"expires": "1 hour",
	})
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	bucket := c.Param("bucket")
	objectName := c.Param("object")

	ctx := context.Background()
	err := h.MinioClient.DeleteFile(ctx, bucket, objectName)
	if err != nil {
		response.InternalError(c, "Failed to delete file")
		return
	}

	response.SuccessWithMessage(c, "File deleted successfully", nil)
}
```

### 6. Inject Minio vào API Handler

**File**: `src/api/handler.go`

```go
type APIHandler struct {
	UserClient     *client.GRPCUser
	FileClient     *client.GRPCfile
	MinioClient    *storage.MinioClient  // Thêm Minio
	Config         *config.Config
}

type ClientOption func(*APIHandler)

func WithMinioClient(client *storage.MinioClient, cfg *config.Config) ClientOption {
	return func(h *APIHandler) {
		h.MinioClient = client
		h.Config = cfg
	}
}
```

**File**: `src/router/router.go`

```go
func setupRestAPI(r *gin.Engine, c *container.Container) {
	apiHandler := api.NewAPIHandler(
		api.WithUserClient(c.Clients.User),
		api.WithFileClient(c.Clients.File),
		api.WithMinioClient(c.Storage, c.Config),  // Inject Minio
	)

	apiV1 := r.Group("/api/v1")
	apiHandler.RegisterRoutes(apiV1)
}
```

### 7. Update Routes

**File**: `src/api/handler.go`

```go
func (h *APIHandler) RegisterRoutes(r *gin.RouterGroup) {
	// ... existing routes

	// File routes
	files := r.Group("/files")
	{
		files.POST("/upload", AuthMiddleware(), h.UploadFile)
		files.GET("/url", h.GetFileURL)
		files.DELETE("/:bucket/:object", AuthMiddleware(), h.DeleteFile)
	}
}
```

## Usage Examples

### Upload File

```bash
curl -X POST http://localhost:8081/api/v1/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/file.pdf" \
  -F "bucket=thesis"
```

**Response**:
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "filename": "thesis.pdf",
    "size": 1048576,
    "url": "http://localhost:9000/thesis/uuid.pdf",
    "object_name": "uuid.pdf",
    "bucket": "thesis"
  }
}
```

### Get Pre-signed URL

```bash
curl "http://localhost:8081/api/v1/files/url?bucket=thesis&object=uuid.pdf"
```

**Response**:
```json
{
  "success": true,
  "data": {
    "url": "http://localhost:9000/thesis/uuid.pdf?X-Amz-...",
    "expires": "1 hour"
  }
}
```

### Delete File

```bash
curl -X DELETE http://localhost:8081/api/v1/files/thesis/uuid.pdf \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Best Practices

### 1. File Validation

```go
// Validate file size
if file.Size > maxSize {
    response.BadRequest(c, "File too large")
    return
}

// Validate file type
allowedTypes := map[string]bool{
    "application/pdf": true,
    "image/jpeg":      true,
    "image/png":       true,
}

if !allowedTypes[file.Header.Get("Content-Type")] {
    response.BadRequest(c, "Invalid file type")
    return
}
```

### 2. Organize Files by Date

```go
// Generate object name with date structure
now := time.Now()
objectName := fmt.Sprintf(
    "%d/%02d/%s%s",
    now.Year(),
    now.Month(),
    uuid.New().String(),
    ext,
)
// Result: 2025/10/uuid.pdf
```

### 3. Cleanup Temporary Files

```go
// Delete old temporary files
func (m *MinioClient) CleanupOldFiles(ctx context.Context, bucket string, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	for object := range m.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Recursive: true,
	}) {
		if object.LastModified.Before(cutoff) {
			m.client.RemoveObject(ctx, bucket, object.Key, minio.RemoveObjectOptions{})
		}
	}

	return nil
}
```

### 4. Bucket Lifecycle Policies

```go
// Set expiration policy for temporary files
config := lifecycle.NewConfiguration()
config.Rules = []lifecycle.Rule{
	{
		ID:     "expire-temp-files",
		Status: "Enabled",
		Expiration: lifecycle.Expiration{
			Days: 7,
		},
		Filter: lifecycle.Filter{
			Prefix: "temp/",
		},
	},
}

err := m.client.SetBucketLifecycle(ctx, "documents", config)
```

## Monitoring

### Check Minio Health

```bash
curl http://localhost:9000/minio/health/live
```

### Monitor Storage Usage

```go
func (m *MinioClient) GetStorageInfo(ctx context.Context) error {
	info, err := m.client.ServerInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Total Space: %d\n", info.Data.Usage.Size)
	return nil
}
```

## Troubleshooting

### Issue: Connection refused

```bash
# Check if Minio is running
docker ps | grep minio

# Check Minio logs
docker logs minio
```

### Issue: Access denied

```bash
# Verify credentials
mc admin info myminio

# Check bucket policy
mc anonymous get myminio/bucket-name
```

### Issue: Bucket not found

```bash
# List all buckets
mc ls myminio/

# Create bucket if missing
mc mb myminio/new-bucket
```
