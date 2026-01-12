package financial

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *service {
	return &service{
		repo:   repo,
		logger: logger,
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
		ID:        uuid.New(),
		Date:      date,
		Amount:    req.Amount,
		Type:      req.Type,
		CreatedAt: now,
		UpdatedAt: now,
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

