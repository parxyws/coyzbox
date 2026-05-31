package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/parxyws/cozybox/internal/config"
)

// InitMinio initializes a global AWS S3 / MinIO client
func InitMinio(cfg *config.Config) (*minio.Client, error) {
	// 1. MinIO Initialization
	minioClient, err := minio.New(cfg.AWS.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AWS.MinioAccessKey, cfg.AWS.MinioSecretKey, ""),
		Secure: cfg.AWS.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// 2. Verification Ping (Optional, but best practice during setup)
	// We can check connectivity by listing buckets or using a simple health check
	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		// Log the warning but don't strictly crash the app if the internet is just down momentarily
		log.Printf("warning: created MinIO client but failed to ping server: %v", err)
	}

	return minioClient, nil
}
