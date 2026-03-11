package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const presignExpiry = 10 * time.Minute

// StorageService abstracts file storage operations.
type StorageService interface {
	// GeneratePresignedURL returns a presigned PUT URL and the resulting object key.
	GeneratePresignedURL(ctx context.Context, filename, contentType, folder string) (presignedURL, objectKey string, err error)
	// DeleteFile removes an object by its key.
	DeleteFile(ctx context.Context, objectKey string) error
	// PublicURL returns the CDN URL for a given object key.
	PublicURL(objectKey string) string
}

type S3StorageService struct {
	client       *s3.Client
	presigner    *s3.PresignClient
	bucketName   string
	cdnDomain    string
}

func NewS3StorageService(cfg config.Config) (StorageService, error) {
	if cfg.S3BucketName == "" {
		log.Warn().Msg("S3_BUCKET_NAME is not set — storage operations will fail")
		return &S3StorageService{bucketName: ""}, nil
	}

	region := cfg.S3Region
	if region == "" {
		region = "ap-southeast-1"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3StorageService{
		client:     client,
		presigner:  s3.NewPresignClient(client),
		bucketName: cfg.S3BucketName,
		cdnDomain:  strings.TrimRight(cfg.CloudFrontDomain, "/"),
	}, nil
}

func (s *S3StorageService) GeneratePresignedURL(ctx context.Context, filename, contentType, folder string) (string, string, error) {
	if s.client == nil {
		return "", "", fmt.Errorf("S3 storage client is not configured")
	}

	objectKey := fmt.Sprintf("%s/%s_%s", folder, uuid.New().String(), filename)

	out, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucketName,
		Key:         &objectKey,
		ContentType: &contentType,
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return out.URL, objectKey, nil
}

func (s *S3StorageService) DeleteFile(ctx context.Context, objectKey string) error {
	if s.client == nil || objectKey == "" {
		return nil
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		log.Warn().Err(err).Str("key", objectKey).Msg("Failed to delete S3 object")
	}
	return nil
}

func (s *S3StorageService) PublicURL(objectKey string) string {
	if s.cdnDomain == "" {
		return objectKey
	}
	return fmt.Sprintf("https://%s/%s", s.cdnDomain, objectKey)
}
