package financial

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*Transaction, error)
	ListTransactions(ctx context.Context, limit, offset int) ([]*Transaction, int64, error)
	GetMonthlyAggregate(ctx context.Context, month string) (*AggregatedData, error)
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	transaction, err := h.service.CreateTransaction(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, transaction)
}

func (h *Handler) ListTransactions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	transactions, total, err := h.service.ListTransactions(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to list transactions"})
		return
	}

	response := ListTransactionsResponse{
		Transactions: transactions,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	c.JSON(200, response)
}

func (h *Handler) GetMonthlyAggregate(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(400, gin.H{"error": "month query parameter is required (format: YYYY-MM)"})
		return
	}

	aggregate, err := h.service.GetMonthlyAggregate(c.Request.Context(), month)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, aggregate)
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(400, gin.H{"error": "transaction ID is required"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid transaction ID"})
		return
	}

	if err := h.service.DeleteTransaction(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete transaction",
			slog.String("error", err.Error()),
			slog.String("id", id.String()))
		c.JSON(500, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.Status(204)
}

