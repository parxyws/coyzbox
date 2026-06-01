package aws

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/parxyws/cozybox/internal/config"
)

type UploadInput struct {
	Object      io.Reader
	ObjectName  string
	ObjectSize  int64
	BucketName  string
	ContentType string
}

type S3Service struct {
	s3Client *minio.Client
	cfg      *config.Config
}

func NewAWSService(s3Client *minio.Client, cfg *config.Config) *S3Service {
	return &S3Service{s3Client: s3Client, cfg: cfg}
}

func (aws *S3Service) PutObject(ctx context.Context, entity *UploadInput) (*minio.UploadInfo, error) {
	opts := minio.PutObjectOptions{
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		ContentType:  entity.ContentType,
	}

	uploadInfo, err := aws.s3Client.PutObject(ctx, entity.BucketName, entity.ObjectName, entity.Object, entity.ObjectSize, opts)
	if err != nil {
		return nil, fmt.Errorf("AWSUserRepository.PutObject.s3Client.PutObject - %s", err)
	}
	return &uploadInfo, nil
}

func (aws *S3Service) GetObject(ctx context.Context, bucketName string, objectName string) (*minio.Object, error) {
	object, err := aws.s3Client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("AWSUserRepository.PutObject.s3Client.GetObject - %s", err)
	}
	defer func(object *minio.Object) {
		err := object.Close()
		if err != nil {
			return
		}
	}(object)

	return object, err
}

func (aws *S3Service) RemoveObject(ctx context.Context, bucketName string, objectName string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	err := aws.s3Client.RemoveObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return fmt.Errorf("AWSUserRepository.PutObject.s3Client.RemoveObject - %s", err)
	}

	return nil
}

func (aws *S3Service) PresignedGetObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (*url.URL, error) {
	var reqParam = make(url.Values)

	presignedUrl, err := aws.s3Client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParam)
	if err != nil {
		return nil, fmt.Errorf("AWSUserRepository.PutObject.s3Client.PresignedGetObject - %s", err)
	}

	return presignedUrl, nil
}

func (aws *S3Service) GenerateAWSMinioURL(bucket string, key string) string {
	return fmt.Sprintf("%s/%s/%s", aws.cfg.AWS.Endpoint, bucket, key)
}
