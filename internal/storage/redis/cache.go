package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Cache struct {
	client *Client
}

func NewCache(client *Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	data, err := c.client.GetDB().Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("cache get failed: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal failed: %w", err)
	}

	return nil
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}

	if err := c.client.GetDB().Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set failed: %w", err)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.GetDB().Del(ctx, key).Err()
}

func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.GetDB().Exists(ctx, key).Result()
	return n > 0, err
}

func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.GetDB().Incr(ctx, key).Result()
}