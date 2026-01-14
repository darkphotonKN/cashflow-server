package config

import (
	"database/sql"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kranti/cashflow/internal/financial"
	"github.com/kranti/cashflow/internal/middleware"
	"github.com/kranti/cashflow/internal/s3"
	"github.com/kranti/cashflow/internal/upload"
)

func SetupRoutes(db *sql.DB, s3Service s3.Service, logger *slog.Logger) *gin.Engine {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.StructuredLogger(logger))
	router.Use(corsMiddleware())

	// Initialize upload services
	uploadRepo := upload.NewRepository(db)
	uploadService := upload.NewService(uploadRepo, s3Service, logger)
	uploadHandler := upload.NewHandler(uploadService, logger)

	// Initialize financial services with upload service dependency
	financialRepo := financial.NewRepository(db)
	financialService := financial.NewService(financialRepo, s3Service, uploadService, logger)
	financialHandler := financial.NewHandler(financialService, logger)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Upload endpoints
		uploads := api.Group("/uploads")
		{
			uploads.POST("/request", uploadHandler.RequestUpload)
			uploads.GET("/:id/status", uploadHandler.GetUploadStatus)
		}

		// Transaction endpoints
		transactions := api.Group("/transactions")
		{
			transactions.POST("", financialHandler.CreateTransaction)
			transactions.GET("", financialHandler.ListTransactions)
			transactions.GET("/aggregate", financialHandler.GetMonthlyAggregate)
			transactions.DELETE("/:id", financialHandler.DeleteTransaction)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Content-Type", "Authorization"}
	return cors.New(config)
}