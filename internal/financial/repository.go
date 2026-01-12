package financial

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, transaction *Transaction) error
	List(ctx context.Context, limit, offset int) ([]*Transaction, error)
	Count(ctx context.Context) (int64, error)
	GetByMonth(ctx context.Context, year int, month int) ([]*Transaction, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, transaction *Transaction) error {
	query := `
		INSERT INTO transactions (id, date, amount, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID,
		transaction.Date,
		transaction.Amount,
		transaction.Type,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("creating transaction: %w", err)
	}

	return nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*Transaction, error) {
	query := `
		SELECT id, date, amount, type, created_at, updated_at
		FROM transactions
		ORDER BY date DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var t Transaction
		err := rows.Scan(
			&t.ID,
			&t.Date,
			&t.Amount,
			&t.Type,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning transaction: %w", err)
		}
		transactions = append(transactions, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transactions: %w", err)
	}

	return transactions, nil
}

func (r *repository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM transactions`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting transactions: %w", err)
	}

	return count, nil
}

func (r *repository) GetByMonth(ctx context.Context, year int, month int) ([]*Transaction, error) {
	query := `
		SELECT id, date, amount, type, created_at, updated_at
		FROM transactions
		WHERE EXTRACT(YEAR FROM date) = $1 AND EXTRACT(MONTH FROM date) = $2
		ORDER BY date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, year, month)
	if err != nil {
		return nil, fmt.Errorf("getting transactions by month: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var t Transaction
		err := rows.Scan(
			&t.ID,
			&t.Date,
			&t.Amount,
			&t.Type,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning transaction: %w", err)
		}
		transactions = append(transactions, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transactions: %w", err)
	}

	return transactions, nil
}