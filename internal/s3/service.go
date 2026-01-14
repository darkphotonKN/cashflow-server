package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Service interface {
	UploadImage(ctx context.Context, imageData []byte, contentType string) (url string, key string, err error)
	DeleteImage(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string) (string, error)
	GeneratePresignedPutURL(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)
	ObjectExists(ctx context.Context, key string) (bool, error)
	CopyObject(ctx context.Context, sourceKey string, destKey string) error
}

type service struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	config        *Config
}

func NewService(cfg *Config) (Service, error) {
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig)
	presignClient := s3.NewPresignClient(client)

	return &service{
		client:        client,
		presignClient: presignClient,
		config:        cfg,
	}, nil
}

func (s *service) UploadImage(ctx context.Context, imageData []byte, contentType string) (string, string, error) {
	if int64(len(imageData)) > s.config.MaxImageSize {
		return "", "", fmt.Errorf("image size exceeds maximum allowed size of %d bytes", s.config.MaxImageSize)
	}

	if !isValidContentType(contentType) {
		return "", "", fmt.Errorf("invalid content type: %s", contentType)
	}

	now := time.Now()
	key := fmt.Sprintf("transactions/%d/%02d/%s_%d.jpg",
		now.Year(),
		now.Month(),
		uuid.New().String(),
		now.Unix(),
	)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"upload-time": now.Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("uploading to S3: %w", err)
	}

	url, err := s.GetPresignedURL(ctx, key)
	if err != nil {
		return "", "", fmt.Errorf("generating presigned URL: %w", err)
	}

	return url, key, nil
}

func (s *service) DeleteImage(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting from S3: %w", err)
	}

	return nil
}

func (s *service) GetPresignedURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.config.URLExpiration
	})
	if err != nil {
		return "", fmt.Errorf("creating presigned URL: %w", err)
	}

	return request.URL, nil
}

func (s *service) GeneratePresignedPutURL(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
	request, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("generating presigned PUT URL: %w", err)
	}

	return request.URL, nil
}

func (s *service) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if the error is because the object doesn't exist
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("checking object existence: %w", err)
	}

	return true, nil
}

func (s *service) CopyObject(ctx context.Context, sourceKey string, destKey string) error {
	copySource := fmt.Sprintf("%s/%s", s.config.BucketName, sourceKey)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.BucketName),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
	})

	if err != nil {
		return fmt.Errorf("copying S3 object: %w", err)
	}

	return nil
}

func isValidContentType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

