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
			"service": "portfolio-service",
		})
	})

	// Portfolio endpoints placeholder
	r.GET("/api/v1/portfolio", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Portfolio Service",
			"version": "0.1.0",
		})
	})

	log.Println("Starting Portfolio Service on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}