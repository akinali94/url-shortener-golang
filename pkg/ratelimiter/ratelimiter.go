package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client      *redis.Client
	keyPrefix   string
	defaultRate Rate
}

type Rate struct {
	Limit  int
	Window time.Duration
}

type Options struct {
	Redis       *redis.Options
	KeyPrefix   string
	DefaultRate Rate
}

func NewRateLimiter(opts Options) (*RateLimiter, error) {
	if opts.Redis == nil {
		opts.Redis = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	client := redis.NewClient(opts.Redis)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if opts.KeyPrefix == "" {
		opts.KeyPrefix = "ratelimit:"
	}

	if opts.DefaultRate.Limit == 0 {
		opts.DefaultRate.Limit = 100
	}

	if opts.DefaultRate.Window == 0 {
		opts.DefaultRate.Window = time.Minute
	}

	return &RateLimiter{
		client:      client,
		keyPrefix:   opts.KeyPrefix,
		defaultRate: opts.DefaultRate,
	}, nil
}

// Allow checks if a request is allowed based on the rate limit
func (rl *RateLimiter) Allow(ctx context.Context, key string, rate ...Rate) (bool, int, error) {
	// Use default rate if none provided
	r := rl.defaultRate
	if len(rate) > 0 {
		r = rate[0]
	}

	// Format Redis key
	redisKey := fmt.Sprintf("%s%s", rl.keyPrefix, key)

	// Current timestamp in milliseconds
	now := time.Now().UnixNano() / int64(time.Millisecond)
	windowMs := int64(r.Window / time.Millisecond)

	// Define the window start (current time minus window duration)
	windowStart := now - windowMs

	// Use Redis pipeline to run multiple commands efficiently
	pipe := rl.client.Pipeline()

	// Remove old entries outside current window
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// Count entries in current window
	countCmd := pipe.ZCount(ctx, redisKey, fmt.Sprintf("%d", windowStart), "+inf")

	// Add current request with score as current timestamp
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	// Set expiration on the key to save memory (window duration + buffer)
	pipe.Expire(ctx, redisKey, r.Window+time.Minute)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("rate limiter Redis error: %w", err)
	}

	// Get current count
	count, err := countCmd.Result()
	if err != nil {
		return false, 0, fmt.Errorf("rate limiter count error: %w", err)
	}

	// Check if we're under the limit
	allowed := count < int64(r.Limit)
	remaining := r.Limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return allowed, remaining, nil
}

// Close releases Redis resources
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}
