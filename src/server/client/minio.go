package client

import (
	"context"
	"fmt"
	"log"
	"thaily/src/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ServiceMinIo struct {
	client *minio.Client
	config *config.MinioConfig
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
