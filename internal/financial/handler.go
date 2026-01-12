package financial

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*Transaction, error)
	ListTransactions(ctx context.Context, limit, offset int) ([]*Transaction, int64, error)
	GetMonthlyAggregate(ctx context.Context, month string) (*AggregatedData, error)
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", slog.String("error", err.Error()))
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	transaction, err := h.service.CreateTransaction(r.Context(), req)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusCreated, transaction)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	transactions, total, err := h.service.ListTransactions(r.Context(), limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list transactions")
		return
	}

	response := ListTransactionsResponse{
		Transactions: transactions,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handler) GetMonthlyAggregate(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		h.respondWithError(w, http.StatusBadRequest, "month query parameter is required (format: YYYY-MM)")
		return
	}

	aggregate, err := h.service.GetMonthlyAggregate(r.Context(), month)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, aggregate)
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}