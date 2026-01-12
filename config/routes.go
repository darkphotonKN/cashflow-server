package config

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/kranti/cashflow/internal/financial"
	"github.com/kranti/cashflow/internal/middleware"
)

func SetupRoutes(db *sql.DB, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	repo := financial.NewRepository(db)
	service := financial.NewService(repo, logger)
	handler := financial.NewHandler(service, logger)

	mux.HandleFunc("POST /api/transactions", handler.CreateTransaction)
	mux.HandleFunc("GET /api/transactions", handler.ListTransactions)
	mux.HandleFunc("GET /api/transactions/aggregate", handler.GetMonthlyAggregate)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return middleware.CORS(mux)
}