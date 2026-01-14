package upload

import (
	"time"

	"github.com/google/uuid"
)

type UploadStatus string

const (
	UploadStatusPending   UploadStatus = "pending"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusFailed    UploadStatus = "failed"
	UploadStatusExpired   UploadStatus = "expired"
)

type UploadRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,min=1,max=10485760"` // Max 10MB
}

type UploadResponse struct {
	UploadID     string            `json:"upload_id"`
	PresignedURL string            `json:"presigned_url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers,omitempty"`
	Key          string            `json:"key"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

type UploadRecord struct {
	ID                     uuid.UUID     `json:"id"`
	UploadID               string        `json:"upload_id"`
	S3Key                  string        `json:"s3_key"`
	ContentType            string        `json:"content_type"`
	FileSize               int64         `json:"file_size"`
	Status                 UploadStatus  `json:"status"`
	PresignedURLExpiresAt  time.Time     `json:"presigned_url_expires_at"`
	CreatedAt              time.Time     `json:"created_at"`
	CompletedAt            *time.Time    `json:"completed_at,omitempty"`
	TransactionID          *uuid.UUID    `json:"transaction_id,omitempty"`
}

type UploadStatusResponse struct {
	UploadID    string       `json:"upload_id"`
	Status      UploadStatus `json:"status"`
	S3Key       string       `json:"s3_key"`
	ContentType string       `json:"content_type"`
	FileSize    int64        `json:"file_size"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}