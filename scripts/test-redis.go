package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"hedge-fund/pkg/shared/config"
	"hedge-fund/pkg/shared/logger"
	"hedge-fund/pkg/shared/queue"
	"hedge-fund/pkg/shared/redis"
	"hedge-fund/pkg/shared/models"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.Env); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Test Redis connection
	fmt.Println("🔌 Testing Redis connection...")
	redisClient, err := redis.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Test health check
	fmt.Println("🏥 Testing Redis health check...")
	if err := redisClient.Health(); err != nil {
		log.Fatal("Redis health check failed:", err)
	}
	fmt.Println("✅ Redis health check passed!")

	// Test basic caching
	fmt.Println("📦 Testing basic caching...")
	ctx := context.Background()

	testData := map[string]interface{}{
		"symbol": "AAPL",
		"price":  150.25,
		"time":   time.Now(),
	}

	if err := redisClient.SetCache(ctx, "test:aapl", testData, time.Minute); err != nil {
		log.Fatal("Failed to set cache:", err)
	}

	var retrievedData map[string]interface{}
	if err := redisClient.GetCache(ctx, "test:aapl", &retrievedData); err != nil {
		log.Fatal("Failed to get cache:", err)
	}
	fmt.Printf("✅ Cache test passed! Retrieved: %+v\n", retrievedData)

	// Test job queue
	fmt.Println("⚙️ Testing job queue...")
	queueManager := queue.NewManager(redisClient)

	// Test enqueue
	jobID, err := queueManager.EnqueueAIAnalysis("TSLA", []string{"warren_buffett", "michael_burry"}, 1)
	if err != nil {
		log.Fatal("Failed to enqueue job:", err)
	}
	fmt.Printf("✅ Job enqueued successfully! Job ID: %s\n", jobID)

	// Check queue length
	length, err := queueManager.GetQueueLength(models.QueueAIAnalysis)
	if err != nil {
		log.Fatal("Failed to get queue length:", err)
	}
	fmt.Printf("✅ Queue length: %d\n", length)

	// Test job status
	if err := queueManager.SetJobStatus(jobID, models.JobStatusRunning, "Processing test job", 50.0); err != nil {
		log.Fatal("Failed to set job status:", err)
	}

	status, err := queueManager.GetJobStatus(jobID)
	if err != nil {
		log.Fatal("Failed to get job status:", err)
	}
	fmt.Printf("✅ Job status test passed! Status: %s, Progress: %.1f%%\n", status.Status, status.Progress)

	// Test pub/sub
	fmt.Println("📡 Testing pub/sub...")
	event := models.PriceUpdateEvent{
		Event: models.Event{
			Type:      "price_update",
			Source:    "test",
			Timestamp: time.Now(),
		},
		Symbol: "AAPL",
		Price:  151.00,
		Change: 0.75,
		Volume: 1000000,
	}

	if err := redisClient.PublishEvent(ctx, models.ChannelPriceUpdates, event); err != nil {
		log.Fatal("Failed to publish event:", err)
	}
	fmt.Println("✅ Event published successfully!")

	// Test market data caching
	fmt.Println("📈 Testing market data caching...")
	marketData := models.MarketData{
		Symbol:       "GOOGL",
		CurrentPrice: 142.50,
		Volume:       2500000,
		LastUpdated:  time.Now(),
	}

	if err := redisClient.SetMarketData(ctx, "GOOGL", marketData); err != nil {
		log.Fatal("Failed to cache market data:", err)
	}

	var retrievedMarketData models.MarketData
	if err := redisClient.GetMarketData(ctx, "GOOGL", &retrievedMarketData); err != nil {
		log.Fatal("Failed to retrieve market data:", err)
	}
	fmt.Printf("✅ Market data cache test passed! Symbol: %s, Price: $%.2f\n",
		retrievedMarketData.Symbol, retrievedMarketData.CurrentPrice)

	// Test session storage
	fmt.Println("👤 Testing session storage...")
	sessionData := map[string]interface{}{
		"user_id": 123,
		"username": "testuser",
		"role": "trader",
		"login_time": time.Now(),
	}

	sessionID := "test-session-123"
	if err := redisClient.SetSession(ctx, sessionID, sessionData, time.Hour); err != nil {
		log.Fatal("Failed to set session:", err)
	}

	var retrievedSession map[string]interface{}
	if err := redisClient.GetSession(ctx, sessionID, &retrievedSession); err != nil {
		log.Fatal("Failed to get session:", err)
	}
	fmt.Printf("✅ Session storage test passed! User: %v\n", retrievedSession["username"])

	// Clean up test data
	fmt.Println("🧹 Cleaning up test data...")
	redisClient.DeleteCache(ctx, "test:aapl")
	redisClient.DeleteSession(ctx, sessionID)
	fmt.Println("✅ Cleanup completed!")

	fmt.Println("\n🎉 All Redis tests passed successfully!")
	fmt.Println("Redis is ready for production use!")
}