package redis

import (
	"context"
	"fmt"
	"time"
)

type RateLimitStore struct {
	client *Client
}

func NewRateLimitStore(client *Client) *RateLimitStore {
	return &RateLimitStore{client: client}
}

func (r *RateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)

	pipe := r.client.GetDB().Pipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("rate limit increment failed: %w", err)
	}

	return incr.Val(), nil
}

func (r *RateLimitStore) Reset(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	return r.client.GetDB().Del(ctx, redisKey).Err()
}