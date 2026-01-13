package config

import (
	"database/sql"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kranti/cashflow/internal/financial"
	"github.com/kranti/cashflow/internal/middleware"
	"github.com/kranti/cashflow/internal/s3"
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

	// Initialize services
	repo := financial.NewRepository(db)
	service := financial.NewService(repo, s3Service, logger)
	handler := financial.NewHandler(service, logger)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		transactions := api.Group("/transactions")
		{
			transactions.POST("", handler.CreateTransaction)
			transactions.GET("", handler.ListTransactions)
			transactions.GET("/aggregate", handler.GetMonthlyAggregate)
			transactions.DELETE("/:id", handler.DeleteTransaction)
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