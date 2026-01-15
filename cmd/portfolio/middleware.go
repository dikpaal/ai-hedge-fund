package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hedge-fund/pkg/shared/database"
	"hedge-fund/pkg/shared/logger"
	"hedge-fund/pkg/shared/redis"
)

// corsMiddleware adds CORS headers to all responses
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// loggingMiddleware logs all HTTP requests with structured logging
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		logger.Info("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}

// recoveryMiddleware recovers from panics and returns 500 error
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// errorMiddleware logs errors after handlers execute
func errorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors after handlers execute
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("Request error",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
			)
		}
	}
}

// healthCheckHandler returns the health status of the service
func healthCheckHandler(db *database.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := gin.H{
			"status":  "ok",
			"service": "portfolio-service",
			"time":    time.Now().UTC().Format(time.RFC3339),
		}

		// Check database health
		if err := db.Health(); err != nil {
			health["status"] = "degraded"
			health["database"] = "unhealthy"
			health["database_error"] = err.Error()
			logger.Warn("Database health check failed", zap.Error(err))
		} else {
			health["database"] = "healthy"
		}

		// Check Redis health
		if err := redisClient.Health(); err != nil {
			health["status"] = "degraded"
			health["redis"] = "unhealthy"
			health["redis_error"] = err.Error()
			logger.Warn("Redis health check failed", zap.Error(err))
		} else {
			health["redis"] = "healthy"
		}

		statusCode := http.StatusOK
		if health["status"] == "degraded" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, health)
	}
}
