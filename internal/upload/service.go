package upload

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kranti/cashflow/internal/s3"
)

type service struct {
	repo      Repository
	s3Service s3.Service
	logger    *slog.Logger
}

func NewService(repo Repository, s3Service s3.Service, logger *slog.Logger) *service {
	return &service{
		repo:      repo,
		s3Service: s3Service,
		logger:    logger,
	}
}

func (s *service) RequestUpload(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
	// Validate content type
	if !isValidContentType(req.ContentType) {
		return nil, fmt.Errorf("invalid content type: %s", req.ContentType)
	}

	// Validate file size
	if req.FileSize > 10*1024*1024 { // 10MB
		return nil, fmt.Errorf("file size exceeds maximum of 10MB")
	}

	// Generate unique upload ID
	uploadID := uuid.New().String()

	// Generate S3 key in staging area
	ext := getExtensionFromContentType(req.ContentType)
	now := time.Now()
	s3Key := fmt.Sprintf("staging/%d/%02d/%s_%d%s",
		now.Year(),
		now.Month(),
		uploadID,
		now.Unix(),
		ext,
	)

	// Generate presigned URL for PUT
	expiresIn := 15 * time.Minute
	presignedURL, err := s.s3Service.GeneratePresignedPutURL(ctx, s3Key, req.ContentType, expiresIn)
	if err != nil {
		s.logger.Error("failed to generate presigned URL",
			slog.String("error", err.Error()),
			slog.String("upload_id", uploadID))
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	// Create upload record
	record := &UploadRecord{
		ID:                    uuid.New(),
		UploadID:              uploadID,
		S3Key:                 s3Key,
		ContentType:           req.ContentType,
		FileSize:              req.FileSize,
		Status:                UploadStatusPending,
		PresignedURLExpiresAt: time.Now().Add(expiresIn),
		CreatedAt:             time.Now(),
	}

	if err := s.repo.Create(ctx, record); err != nil {
		s.logger.Error("failed to create upload record",
			slog.String("error", err.Error()),
			slog.String("upload_id", uploadID))
		return nil, fmt.Errorf("creating upload record: %w", err)
	}

	s.logger.Info("upload request created",
		slog.String("upload_id", uploadID),
		slog.String("s3_key", s3Key),
		slog.Int64("file_size", req.FileSize))

	return &UploadResponse{
		UploadID:     uploadID,
		PresignedURL: presignedURL,
		Method:       "PUT",
		Headers: map[string]string{
			"Content-Type": req.ContentType,
		},
		Key:       s3Key,
		ExpiresAt: record.PresignedURLExpiresAt,
	}, nil
}

func (s *service) GetUploadStatus(ctx context.Context, uploadID string) (*UploadStatusResponse, error) {
	record, err := s.repo.GetByUploadID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("getting upload record: %w", err)
	}

	// Check if upload actually exists in S3 if status is pending
	if record.Status == UploadStatusPending {
		exists, err := s.s3Service.ObjectExists(ctx, record.S3Key)
		if err != nil {
			s.logger.Error("failed to check S3 object",
				slog.String("error", err.Error()),
				slog.String("upload_id", uploadID))
		} else if exists {
			// Update status to completed if object exists
			if err := s.repo.UpdateStatus(ctx, uploadID, UploadStatusCompleted); err != nil {
				s.logger.Error("failed to update upload status",
					slog.String("error", err.Error()),
					slog.String("upload_id", uploadID))
			} else {
				record.Status = UploadStatusCompleted
			}
		}
	}

	return &UploadStatusResponse{
		UploadID:    record.UploadID,
		Status:      record.Status,
		S3Key:       record.S3Key,
		ContentType: record.ContentType,
		FileSize:    record.FileSize,
		CreatedAt:   record.CreatedAt,
		CompletedAt: record.CompletedAt,
	}, nil
}

func (s *service) VerifyAndLinkUpload(ctx context.Context, uploadID string, transactionID uuid.UUID) (string, error) {
	if uploadID == "" {
		return "", nil // No upload to verify
	}

	// Get upload record
	record, err := s.repo.GetByUploadID(ctx, uploadID)
	if err != nil {
		return "", fmt.Errorf("getting upload record: %w", err)
	}

	// Check if already linked
	if record.TransactionID != nil {
		return "", fmt.Errorf("upload already linked to another transaction")
	}

	// Verify object exists in S3
	exists, err := s.s3Service.ObjectExists(ctx, record.S3Key)
	if err != nil {
		return "", fmt.Errorf("verifying S3 object: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("uploaded file not found in S3")
	}

	// Move from staging to permanent location
	permanentKey := strings.Replace(record.S3Key, "staging/", "transactions/", 1)
	if err := s.s3Service.CopyObject(ctx, record.S3Key, permanentKey); err != nil {
		s.logger.Error("failed to copy S3 object",
			slog.String("error", err.Error()),
			slog.String("from", record.S3Key),
			slog.String("to", permanentKey))
		return "", fmt.Errorf("moving file to permanent storage: %w", err)
	}

	// Delete staging object
	if err := s.s3Service.DeleteImage(ctx, record.S3Key); err != nil {
		s.logger.Warn("failed to delete staging object",
			slog.String("error", err.Error()),
			slog.String("key", record.S3Key))
		// Continue anyway - lifecycle rule will clean it up
	}

	// Link upload to transaction
	if err := s.repo.LinkToTransaction(ctx, uploadID, transactionID); err != nil {
		return "", fmt.Errorf("linking upload to transaction: %w", err)
	}

	s.logger.Info("upload verified and linked",
		slog.String("upload_id", uploadID),
		slog.String("transaction_id", transactionID.String()),
		slog.String("s3_key", permanentKey))

	return permanentKey, nil
}

func (s *service) CleanupOrphanedUploads(ctx context.Context) error {
	// Get uploads older than 24 hours without transactions
	orphans, err := s.repo.GetOrphanedUploads(ctx, 24)
	if err != nil {
		return fmt.Errorf("getting orphaned uploads: %w", err)
	}

	for _, orphan := range orphans {
		// Delete from S3
		if err := s.s3Service.DeleteImage(ctx, orphan.S3Key); err != nil {
			s.logger.Warn("failed to delete orphaned S3 object",
				slog.String("error", err.Error()),
				slog.String("key", orphan.S3Key))
		}

		// Update status to expired
		if err := s.repo.UpdateStatus(ctx, orphan.UploadID, UploadStatusExpired); err != nil {
			s.logger.Warn("failed to update orphan status",
				slog.String("error", err.Error()),
				slog.String("upload_id", orphan.UploadID))
		}
	}

	s.logger.Info("cleaned up orphaned uploads",
		slog.Int("count", len(orphans)))

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

func getExtensionFromContentType(contentType string) string {
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/jpg":  ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	if ext, ok := extensions[contentType]; ok {
		return ext
	}

	// Fallback to parsing content type
	parts := strings.Split(contentType, "/")
	if len(parts) == 2 {
		return "." + parts[1]
	}

	return ".jpg" // Default
}

