package upload

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, record *UploadRecord) error
	GetByUploadID(ctx context.Context, uploadID string) (*UploadRecord, error)
	UpdateStatus(ctx context.Context, uploadID string, status UploadStatus) error
	LinkToTransaction(ctx context.Context, uploadID string, transactionID uuid.UUID) error
	GetOrphanedUploads(ctx context.Context, olderThan int) ([]*UploadRecord, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, record *UploadRecord) error {
	query := `
		INSERT INTO upload_requests (
			id, upload_id, s3_key, content_type, file_size,
			status, presigned_url_expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.UploadID,
		record.S3Key,
		record.ContentType,
		record.FileSize,
		record.Status,
		record.PresignedURLExpiresAt,
		record.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("creating upload record: %w", err)
	}

	return nil
}

func (r *repository) GetByUploadID(ctx context.Context, uploadID string) (*UploadRecord, error) {
	query := `
		SELECT
			id, upload_id, s3_key, content_type, file_size,
			status, presigned_url_expires_at, created_at,
			completed_at, transaction_id
		FROM upload_requests
		WHERE upload_id = $1
	`

	var record UploadRecord
	err := r.db.QueryRowContext(ctx, query, uploadID).Scan(
		&record.ID,
		&record.UploadID,
		&record.S3Key,
		&record.ContentType,
		&record.FileSize,
		&record.Status,
		&record.PresignedURLExpiresAt,
		&record.CreatedAt,
		&record.CompletedAt,
		&record.TransactionID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("upload not found")
		}
		return nil, fmt.Errorf("getting upload record: %w", err)
	}

	return &record, nil
}

func (r *repository) UpdateStatus(ctx context.Context, uploadID string, status UploadStatus) error {
	var query string
	var args []interface{}

	if status == UploadStatusCompleted {
		query = `
			UPDATE upload_requests
			SET status = $1, completed_at = NOW()
			WHERE upload_id = $2
		`
		args = []interface{}{status, uploadID}
	} else {
		query = `
			UPDATE upload_requests
			SET status = $1
			WHERE upload_id = $2
		`
		args = []interface{}{status, uploadID}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating upload status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("upload not found")
	}

	return nil
}

func (r *repository) LinkToTransaction(ctx context.Context, uploadID string, transactionID uuid.UUID) error {
	query := `
		UPDATE upload_requests
		SET transaction_id = $1, status = $2, completed_at = NOW()
		WHERE upload_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, transactionID, UploadStatusCompleted, uploadID)
	if err != nil {
		return fmt.Errorf("linking upload to transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("upload not found")
	}

	return nil
}

func (r *repository) GetOrphanedUploads(ctx context.Context, hoursOld int) ([]*UploadRecord, error) {
	query := `
		SELECT
			id, upload_id, s3_key, content_type, file_size,
			status, presigned_url_expires_at, created_at,
			completed_at, transaction_id
		FROM upload_requests
		WHERE status = $1
		AND transaction_id IS NULL
		AND created_at < NOW() - INTERVAL '%d hours'
	`

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(query, hoursOld), UploadStatusPending)
	if err != nil {
		return nil, fmt.Errorf("getting orphaned uploads: %w", err)
	}
	defer rows.Close()

	var records []*UploadRecord
	for rows.Next() {
		var record UploadRecord
		err := rows.Scan(
			&record.ID,
			&record.UploadID,
			&record.S3Key,
			&record.ContentType,
			&record.FileSize,
			&record.Status,
			&record.PresignedURLExpiresAt,
			&record.CreatedAt,
			&record.CompletedAt,
			&record.TransactionID,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning upload record: %w", err)
		}
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating upload records: %w", err)
	}

	return records, nil
}