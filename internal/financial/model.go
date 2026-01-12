package financial

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeSpending TransactionType = "spending"
	TransactionTypeEarning  TransactionType = "earning"
)

type Transaction struct {
	ID        uuid.UUID       `json:"id"`
	Date      time.Time       `json:"date"`
	Amount    float64         `json:"amount"`
	Type      TransactionType `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateTransactionRequest struct {
	Date   string          `json:"date"`
	Amount float64         `json:"amount"`
	Type   TransactionType `json:"type"`
}

type ListTransactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Total        int64          `json:"total"`
	Limit        int            `json:"limit"`
	Offset       int            `json:"offset"`
}

type AggregatedData struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Spending float64 `json:"spending"`
	NetTotal float64 `json:"net_total"`
}

