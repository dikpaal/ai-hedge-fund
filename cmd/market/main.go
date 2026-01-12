package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "market-data-service",
		})
	})

	// Market data endpoints placeholder
	r.GET("/api/v1/market", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Market Data Service",
			"version": "0.1.0",
		})
	})

	log.Println("Starting Market Data Service on :8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}