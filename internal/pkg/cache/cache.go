package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache provides a high-level caching interface
type Cache struct {
	client *redis.Client
	logger *zap.Logger
	prefix string
}

// Config holds cache configuration
type Config struct {
	Prefix        string
	DefaultTTL    time.Duration
	EnableLogging bool
}

// NewCache creates a new cache instance
func NewCache(client *redis.Client, logger *zap.Logger, config Config) *Cache {
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}
	if config.Prefix == "" {
		config.Prefix = "cache"
	}

	return &Cache{
		client: client,
		logger: logger,
		prefix: config.Prefix,
	}
}

// key generates a cache key with prefix
func (c *Cache) key(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Get retrieves a value from cache and unmarshals it into the destination
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	cacheKey := c.key(key)

	val, err := c.client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// Cache miss
		c.logger.Debug("Cache miss", zap.String("key", key))
		return false, nil
	}
	if err != nil {
		c.logger.Error("Cache get error", zap.String("key", key), zap.Error(err))
		return false, err
	}

	// Unmarshal the cached value
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Error("Cache unmarshal error", zap.String("key", key), zap.Error(err))
		return false, err
	}

	c.logger.Debug("Cache hit", zap.String("key", key))
	return true, nil
}

// Set stores a value in cache with the given TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cacheKey := c.key(key)

	// Marshal the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Cache marshal error", zap.String("key", key), zap.Error(err))
		return err
	}

	// Store in cache
	if err := c.client.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		c.logger.Error("Cache set error", zap.String("key", key), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache set", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// Delete removes a value from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	cacheKey := c.key(key)

	if err := c.client.Del(ctx, cacheKey).Err(); err != nil {
		c.logger.Error("Cache delete error", zap.String("key", key), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache delete", zap.String("key", key))
	return nil
}

// DeletePattern deletes all keys matching a pattern
func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	cachePattern := c.key(pattern)

	// Scan for matching keys
	iter := c.client.Scan(ctx, 0, cachePattern, 0).Iterator()
	keys := []string{}

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		c.logger.Error("Cache scan error", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all matching keys
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error("Cache delete pattern error", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache delete pattern", zap.String("pattern", pattern), zap.Int("count", len(keys)))
	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := c.key(key)

	result, err := c.client.Exists(ctx, cacheKey).Result()
	if err != nil {
		c.logger.Error("Cache exists error", zap.String("key", key), zap.Error(err))
		return false, err
	}

	return result > 0, nil
}

// GetOrSet retrieves a value from cache or computes it using the provided function
func (c *Cache) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, compute func() (interface{}, error)) error {
	// Try to get from cache
	hit, err := c.Get(ctx, key, dest)
	if err != nil {
		return err
	}

	if hit {
		// Cache hit, return
		return nil
	}

	// Cache miss, compute value
	value, err := compute()
	if err != nil {
		return err
	}

	// Store in cache
	if err := c.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail the operation
		c.logger.Warn("Failed to cache computed value", zap.String("key", key), zap.Error(err))
	}

	// Copy computed value to destination
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// TTL returns the remaining time to live of a key
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	cacheKey := c.key(key)

	ttl, err := c.client.TTL(ctx, cacheKey).Result()
	if err != nil {
		c.logger.Error("Cache TTL error", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	return ttl, nil
}

// Refresh extends the TTL of a key without modifying its value
func (c *Cache) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	cacheKey := c.key(key)

	if err := c.client.Expire(ctx, cacheKey, ttl).Err(); err != nil {
		c.logger.Error("Cache refresh error", zap.String("key", key), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache refresh", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// Increment atomically increments a counter
func (c *Cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	cacheKey := c.key(key)

	val, err := c.client.IncrBy(ctx, cacheKey, delta).Result()
	if err != nil {
		c.logger.Error("Cache increment error", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	return val, nil
}

// Decrement atomically decrements a counter
func (c *Cache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	cacheKey := c.key(key)

	val, err := c.client.DecrBy(ctx, cacheKey, delta).Result()
	if err != nil {
		c.logger.Error("Cache decrement error", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	return val, nil
}
