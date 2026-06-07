package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
)

func InitMinio(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.AWS.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AWS.MinioAccessKey, cfg.AWS.MinioSecretKey, ""),
		Secure: cfg.AWS.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		log.Printf("warning: created MinIO client but failed to ping server: %v", err)
	}

	return minioClient, nil
}

type S3Service struct {
	client *minio.Client
	cfg    *config.Config
}

func NewS3Service(client *minio.Client, cfg *config.Config) *S3Service {
	return &S3Service{client: client, cfg: cfg}
}

func (s *S3Service) PutObject(ctx context.Context, input domain.UploadInput) (string, error) {
	opts := minio.PutObjectOptions{
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		ContentType:  input.ContentType,
	}

	bucketName := s.cfg.AWS.BucketName
	if bucketName == "" {
		bucketName = "cozybox"
	}
	uploadInfo, err := s.client.PutObject(ctx, bucketName, input.ObjectName, input.Object, input.ObjectSize, opts)
	if err != nil {
		return "", fmt.Errorf("S3Service.PutObject: %w", err)
	}
	return uploadInfo.Key, nil
}

func (s *S3Service) GetObject(ctx context.Context, bucketName string, objectName string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("S3Service.GetObject: %w", err)
	}
	return object, nil
}

func (s *S3Service) RemoveObject(ctx context.Context, bucketName string, objectName string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	if err := s.client.RemoveObject(ctx, bucketName, objectName, opts); err != nil {
		return fmt.Errorf("S3Service.RemoveObject: %w", err)
	}
	return nil
}

func (s *S3Service) GenerateURL(bucket string, key string) string {
	return fmt.Sprintf("%s/%s/%s", s.cfg.AWS.Endpoint, bucket, key)
}
