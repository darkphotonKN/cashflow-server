package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kranti/cashflow/config"
	"github.com/kranti/cashflow/internal/s3"
)

func main() {
	_ = godotenv.Load()

	logLevel := slog.LevelInfo
	if level := os.Getenv("LOG_LEVEL"); level == "debug" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	db, err := config.NewDatabase(logger)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	s3Config, err := s3.NewConfig()
	if err != nil {
		logger.Error("failed to load S3 config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	s3Service, err := s3.NewService(s3Config)
	if err != nil {
		logger.Error("failed to create S3 service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	router := config.SetupRoutes(db, s3Service, logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		logger.Info("starting server", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server shutdown complete")
}

