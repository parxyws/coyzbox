package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
)

func InitMinio(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.MinioAccessKey, cfg.Minio.MinioSecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "warning: created MinIO client but failed to ping server: %v\n", err)
		if err != nil {
			return nil, err
		}
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

	bucketName := s.cfg.Minio.BucketName
	if bucketName == "" {
		bucketName = "cozybox"
	}
	uploadInfo, err := s.client.PutObject(ctx, bucketName, input.ObjectName, input.Object, input.ObjectSize, opts)
	if err != nil {
		return "", fmt.Errorf("S3Service.PutObject: %w", err)
	}
	return uploadInfo.Key, nil
}

// ReadObject retrieves an object from the configured S3-compatible storage
// and passes it to the provided callback function. The object is automatically
// closed after the callback returns, preventing resource leaks.
func (s *S3Service) ReadObject(ctx context.Context, bucketName string, objectName string, fn func(obj *minio.Object) error) error {
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("S3Service.ReadObject: %w", err)
	}
	defer func() {
		if closeErr := object.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close S3 object %s/%s: %v\n", bucketName, objectName, closeErr)
		}
	}()

	return fn(object)
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
	return fmt.Sprintf("%s/%s/%s", s.cfg.Minio.Endpoint, bucket, key)
}
