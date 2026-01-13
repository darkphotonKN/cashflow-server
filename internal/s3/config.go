package s3

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Region          string
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
	URLExpiration   time.Duration
	MaxImageSize    int64
}

func NewConfig() (*Config, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is required")
	}

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID environment variable is required")
	}

	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return nil, fmt.Errorf("AWS_SECRET_ACCESS_KEY environment variable is required")
	}

	urlExpiration := 24 * time.Hour
	if exp := os.Getenv("S3_URL_EXPIRATION"); exp != "" {
		duration, err := time.ParseDuration(exp)
		if err == nil {
			urlExpiration = duration
		}
	}

	maxImageSize := int64(10 * 1024 * 1024) // 10MB default
	if sizeStr := os.Getenv("MAX_IMAGE_SIZE"); sizeStr != "" {
		var size int64
		_, err := fmt.Sscanf(sizeStr, "%d", &size)
		if err == nil && size > 0 {
			maxImageSize = size
		}
	}

	return &Config{
		Region:          region,
		BucketName:      bucketName,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		URLExpiration:   urlExpiration,
		MaxImageSize:    maxImageSize,
	}, nil
}