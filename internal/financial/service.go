package financial

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
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

func (s *service) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*Transaction, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if req.Type != TransactionTypeSpending && req.Type != TransactionTypeEarning {
		return nil, fmt.Errorf("invalid transaction type: %s", req.Type)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	now := time.Now()
	transaction := &Transaction{
		ID:          uuid.New(),
		Date:        date,
		Amount:      req.Amount,
		Type:        req.Type,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Handle image upload if provided
	if req.ImageBase64 != "" {
		imageData, contentType, err := s.decodeBase64Image(req.ImageBase64)
		if err != nil {
			return nil, fmt.Errorf("decoding image: %w", err)
		}

		url, key, err := s.s3Service.UploadImage(ctx, imageData, contentType)
		if err != nil {
			return nil, fmt.Errorf("uploading image: %w", err)
		}

		transaction.ImageURL = url
		transaction.ImageKey = key
	}

	if err := s.repo.Create(ctx, transaction); err != nil {
		s.logger.Error("failed to create transaction",
			slog.String("error", err.Error()),
			slog.String("type", string(req.Type)),
			slog.Float64("amount", req.Amount))
		return nil, fmt.Errorf("creating transaction: %w", err)
	}

	s.logger.Info("transaction created",
		slog.String("id", transaction.ID.String()),
		slog.String("type", string(transaction.Type)),
		slog.Float64("amount", transaction.Amount))

	return transaction, nil
}

func (s *service) ListTransactions(ctx context.Context, limit, offset int) ([]*Transaction, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	transactions, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("failed to list transactions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("listing transactions: %w", err)
	}

	count, err := s.repo.Count(ctx)
	if err != nil {
		s.logger.Error("failed to count transactions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	return transactions, count, nil
}

func (s *service) GetMonthlyAggregate(ctx context.Context, month string) (*AggregatedData, error) {
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid month format, expected YYYY-MM")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}

	monthNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid month: %w", err)
	}

	if monthNum < 1 || monthNum > 12 {
		return nil, fmt.Errorf("month must be between 1 and 12")
	}

	transactions, err := s.repo.GetByMonth(ctx, year, monthNum)
	if err != nil {
		s.logger.Error("failed to get monthly transactions",
			slog.String("error", err.Error()),
			slog.String("month", month))
		return nil, fmt.Errorf("getting monthly transactions: %w", err)
	}

	var income, spending float64
	for _, t := range transactions {
		switch t.Type {
		case TransactionTypeEarning:
			income += t.Amount
		case TransactionTypeSpending:
			spending += t.Amount
		}
	}

	aggregate := &AggregatedData{
		Month:    month,
		Income:   income,
		Spending: spending,
		NetTotal: income - spending,
	}

	s.logger.Info("calculated monthly aggregate",
		slog.String("month", month),
		slog.Float64("income", income),
		slog.Float64("spending", spending),
		slog.Float64("net", aggregate.NetTotal))

	return aggregate, nil
}

func (s *service) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	// Get transaction to retrieve image key
	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting transaction: %w", err)
	}

	// Delete image from S3 if exists
	if transaction.ImageKey != "" {
		if err := s.s3Service.DeleteImage(ctx, transaction.ImageKey); err != nil {
			s.logger.Error("failed to delete image from S3",
				slog.String("error", err.Error()),
				slog.String("key", transaction.ImageKey))
			// Continue with transaction deletion even if image deletion fails
		}
	}

	// Delete transaction from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting transaction: %w", err)
	}

	s.logger.Info("transaction deleted",
		slog.String("id", id.String()))

	return nil
}

func (s *service) decodeBase64Image(base64Str string) ([]byte, string, error) {
	// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
	parts := strings.Split(base64Str, ",")
	var data string
	var contentType string

	if len(parts) == 2 && strings.HasPrefix(parts[0], "data:") {
		// Extract content type from data URL
		metadata := parts[0]
		data = parts[1]

		// Parse content type from metadata
		metaParts := strings.Split(metadata, ":")
		if len(metaParts) == 2 {
			contentParts := strings.Split(metaParts[1], ";")
			if len(contentParts) > 0 {
				contentType = contentParts[0]
			}
		}
	} else {
		data = base64Str
		contentType = "image/jpeg" // default
	}

	imageData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", fmt.Errorf("decoding base64: %w", err)
	}

	return imageData, contentType, nil
}
