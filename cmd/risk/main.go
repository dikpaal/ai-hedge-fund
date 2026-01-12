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
			"service": "risk-service",
		})
	})

	// Risk endpoints placeholder
	r.GET("/api/v1/risk", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Risk Management Service",
			"version": "0.1.0",
		})
	})

	log.Println("Starting Risk Service on :8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}