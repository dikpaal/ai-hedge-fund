package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"hedge-fund/pkg/shared/config"
	"hedge-fund/pkg/shared/logger"
)

type Client struct {
	*redis.Client
}

// Connect establishes a connection to Redis
func Connect(cfg *config.Config) (*Client, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis")

	return &Client{rdb}, nil
}

// Health checks if the Redis connection is healthy
func (c *Client) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Cache operations

// SetCache stores a value in cache with expiration
func (c *Client) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := c.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	logger.Debug("Cache set successfully", zap.String("key", key))
	return nil
}

// GetCache retrieves a value from cache
func (c *Client) GetCache(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache key not found: %s", key)
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	logger.Debug("Cache retrieved successfully", zap.String("key", key))
	return nil
}

// DeleteCache removes a key from cache
func (c *Client) DeleteCache(ctx context.Context, key string) error {
	if err := c.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache key: %w", err)
	}

	logger.Debug("Cache key deleted", zap.String("key", key))
	return nil
}

// CacheExists checks if a cache key exists
func (c *Client) CacheExists(ctx context.Context, key string) (bool, error) {
	count, err := c.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return count > 0, nil
}

// Job Queue operations

// EnqueueJob adds a job to a queue
func (c *Client) EnqueueJob(ctx context.Context, queue string, job interface{}) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := c.LPush(ctx, queue, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	logger.Debug("Job enqueued successfully",
		zap.String("queue", queue),
		zap.Any("job", job))
	return nil
}

// DequeueJob removes and returns a job from a queue (blocking)
func (c *Client) DequeueJob(ctx context.Context, queue string, timeout time.Duration, dest interface{}) error {
	result, err := c.BRPop(ctx, timeout, queue).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("no job available in queue: %s", queue)
		}
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return fmt.Errorf("invalid job result from queue")
	}

	if err := json.Unmarshal([]byte(result[1]), dest); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	logger.Debug("Job dequeued successfully", zap.String("queue", queue))
	return nil
}

// QueueLength returns the number of jobs in a queue
func (c *Client) QueueLength(ctx context.Context, queue string) (int64, error) {
	length, err := c.LLen(ctx, queue).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}

	return length, nil
}

// Session storage operations

// SetSession stores session data
func (c *Client) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.SetCache(ctx, key, data, expiration)
}

// GetSession retrieves session data
func (c *Client) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.GetCache(ctx, key, dest)
}

// DeleteSession removes session data
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.DeleteCache(ctx, key)
}

// Market data caching operations

// SetMarketData caches market data with appropriate TTL
func (c *Client) SetMarketData(ctx context.Context, symbol string, data interface{}) error {
	key := fmt.Sprintf("market:%s", symbol)
	// Market data expires after 1 minute for real-time updates
	return c.SetCache(ctx, key, data, time.Minute)
}

// GetMarketData retrieves cached market data
func (c *Client) GetMarketData(ctx context.Context, symbol string, dest interface{}) error {
	key := fmt.Sprintf("market:%s", symbol)
	return c.GetCache(ctx, key, dest)
}

// SetPriceAlert sets a price alert for a symbol
func (c *Client) SetPriceAlert(ctx context.Context, userID int, symbol string, price float64) error {
	key := fmt.Sprintf("alert:%d:%s", userID, symbol)
	alertData := map[string]interface{}{
		"user_id": userID,
		"symbol":  symbol,
		"price":   price,
		"created": time.Now(),
	}
	// Price alerts don't expire automatically
	return c.SetCache(ctx, key, alertData, 0)
}

// GetPriceAlerts retrieves all price alerts for a user
func (c *Client) GetPriceAlerts(ctx context.Context, userID int) ([]map[string]interface{}, error) {
	pattern := fmt.Sprintf("alert:%d:*", userID)
	keys, err := c.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get alert keys: %w", err)
	}

	var alerts []map[string]interface{}
	for _, key := range keys {
		var alert map[string]interface{}
		if err := c.GetCache(ctx, key, &alert); err == nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// Pub/Sub operations for real-time updates

// PublishEvent publishes an event to a channel
func (c *Client) PublishEvent(ctx context.Context, channel string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := c.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	logger.Debug("Event published successfully",
		zap.String("channel", channel),
		zap.Any("event", event))
	return nil
}

// SubscribeToEvents subscribes to events on a channel
func (c *Client) SubscribeToEvents(ctx context.Context, channel string) *redis.PubSub {
	logger.Info("Subscribing to events", zap.String("channel", channel))
	return c.Subscribe(ctx, channel)
}

// Utility functions

// FlushCache clears all cache data (use with caution)
func (c *Client) FlushCache(ctx context.Context) error {
	if err := c.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush cache: %w", err)
	}

	logger.Warn("Cache flushed - all data cleared")
	return nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	logger.Info("Closing Redis connection")
	return c.Client.Close()
}