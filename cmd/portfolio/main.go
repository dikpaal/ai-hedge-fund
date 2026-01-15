package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hedge-fund/internal/portfolio/domain"
	"hedge-fund/internal/portfolio/handlers"
	"hedge-fund/internal/portfolio/repository"
	"hedge-fund/internal/portfolio/service"
	"hedge-fund/pkg/shared/config"
	"hedge-fund/pkg/shared/database"
	"hedge-fund/pkg/shared/logger"
	"hedge-fund/pkg/shared/redis"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.Env); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting Portfolio Service",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.PortfolioServicePort),
	)

	// Connect to PostgreSQL database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Verify database health
	if err := db.Health(); err != nil {
		logger.Fatal("Database health check failed", zap.Error(err))
	}
	logger.Info("Database connection established")

	// Connect to Redis
	redisClient, err := redis.Connect(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Verify Redis health
	if err := redisClient.Health(); err != nil {
		logger.Fatal("Redis health check failed", zap.Error(err))
	}
	logger.Info("Redis connection established")

	// Create dependency chain
	// Repository layer (database operations)
	portfolioRepo := repository.NewPortfolioRepository(db, logger.Logger)

	// Domain service (business logic)
	domainService := domain.NewPortfolioService()

	// Service layer (orchestration + transactions)
	portfolioService := service.NewPortfolioService(portfolioRepo, domainService, logger.Logger)

	// Mock market client (will be replaced with real Market Data Service later)
	marketClient := handlers.NewMockMarketDataClient()

	// Handler (HTTP layer)
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService, marketClient, logger.Logger)

	// Setup Gin router
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New() // Use New() instead of Default() to have full control over middleware

	// Apply middleware stack (order matters!)
	router.Use(corsMiddleware())      // 1. CORS
	router.Use(loggingMiddleware())   // 2. Request logging
	router.Use(recoveryMiddleware())  // 3. Panic recovery
	router.Use(errorMiddleware())     // 4. Error handling

	// Health check endpoint (outside API versioning)
	router.GET("/health", healthCheckHandler(db, redisClient))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Portfolio CRUD operations
		v1.POST("/portfolios", portfolioHandler.CreatePortfolio)
		v1.GET("/portfolios/:id", portfolioHandler.GetPortfolio)
		v1.PUT("/portfolios/:id", portfolioHandler.UpdatePortfolio)
		v1.DELETE("/portfolios/:id", portfolioHandler.DeletePortfolio)
		v1.GET("/portfolios/user/:user_id", portfolioHandler.ListUserPortfolios)

		// Position operations
		v1.GET("/portfolios/:id/positions", portfolioHandler.GetPositions)

		// Portfolio analysis
		v1.GET("/portfolios/:id/summary", portfolioHandler.GetSummary)
		v1.GET("/portfolios/:id/allocation", portfolioHandler.GetAllocation)
		v1.GET("/portfolios/:id/risk", portfolioHandler.GetRiskMetrics)

		// Trading operations
		v1.POST("/portfolios/:id/trades", portfolioHandler.ExecuteTrade)
		v1.GET("/portfolios/:id/trades", portfolioHandler.GetTradeHistory)

		// Rebalancing
		v1.POST("/portfolios/:id/rebalance", portfolioHandler.GetRebalanceRecommendations)
	}

	// Configure HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.PortfolioServicePort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Portfolio Service listening", zap.String("port", cfg.PortfolioServicePort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Portfolio Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Portfolio Service stopped")
}
