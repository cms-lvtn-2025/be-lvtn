package client

import (
	"context"
	"fmt"
	"log"
	"thaily/src/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

type ServiceMinIo struct {
	client      *minio.Client
	config      *config.MinioConfig
	redisClient *redis.Client
}

func (s *ServiceMinIo) ensureBucket() error {
	ctx := context.Background()
	bucketName := s.config.BucketName

	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Bucket '%s' created successfully", bucketName)
	}

	return nil
}

func NewServiceMinIo(cfg config.MinioConfig) (*ServiceMinIo, error) {
	client, err := minio.New(cfg.Endpont, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	service := &ServiceMinIo{
		client: client,
		config: &cfg,
	}

	// Create bucket if not exists
	if err := service.ensureBucket(); err != nil {
		return nil, err
	}

	return service, nil
}

// UploadFile uploads a file to MinIO
func (s *ServiceMinIo) UploadFile(ctx context.Context, objectName string, reader interface{}, objectSize int64, contentType string) (string, error) {
	bucketName := s.config.BucketName

	_, err := s.client.PutObject(ctx, bucketName, objectName, reader.(interface {
		Read(p []byte) (n int, err error)
		Seek(offset int64, whence int) (int64, error)
	}), objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate file URL
	fileURL := fmt.Sprintf("%s/%s/%s", s.config.Endpont, bucketName, objectName)
	if s.config.UseSSL {
		fileURL = fmt.Sprintf("https://%s/%s/%s", s.config.Endpont, bucketName, objectName)
	} else {
		fileURL = fmt.Sprintf("http://%s/%s/%s", s.config.Endpont, bucketName, objectName)
	}

	log.Printf("File uploaded successfully: %s", fileURL)
	return fileURL, nil
}

// DeleteFile deletes a file from MinIO
func (s *ServiceMinIo) DeleteFile(ctx context.Context, objectName string) error {
	bucketName := s.config.BucketName

	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	log.Printf("File deleted successfully: %s", objectName)
	return nil
}

// GetFileURL generates a presigned URL for file access
func (s *ServiceMinIo) GetFileURL(ctx context.Context, objectName string) (string, error) {
	bucketName := s.config.BucketName

	// Generate presigned URL valid for 7 days
	url, err := s.client.PresignedGetObject(ctx, bucketName, objectName, 7*24*60*60*1000000000, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GetFileBlob retrieves file content directly from MinIO
func (s *ServiceMinIo) GetFileBlob(ctx context.Context, objectName string) (*minio.Object, *minio.ObjectInfo, error) {
	bucketName := s.config.BucketName

	// Get object
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Get object info for content type and size
	objectInfo, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return object, &objectInfo, nil
}
