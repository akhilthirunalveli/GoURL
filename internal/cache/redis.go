package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/models"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr, password string, db int, ttl int) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Cache{
		client: client,
		ttl:    time.Duration(ttl) * time.Second,
	}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Get(ctx context.Context, key string) (*models.URL, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var url models.URL
	if err := json.Unmarshal([]byte(val), &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL: %w", err)
	}

	return &url, nil
}

func (c *Cache) Set(ctx context.Context, key string, url *models.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal URL: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// CheckRateLimit checks if the client has exceeded the rate limit
func (c *Cache) CheckRateLimit(ctx context.Context, clientIP string, limit int, window int) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", clientIP)

	// Increment the counter
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	// Set expiration on first request
	if count == 1 {
		if err := c.client.Expire(ctx, key, time.Duration(window)*time.Second).Err(); err != nil {
			return false, fmt.Errorf("failed to set expiration: %w", err)
		}
	}

	// Check if limit exceeded
	if count > int64(limit) {
		return false, nil
	}

	return true, nil
}
