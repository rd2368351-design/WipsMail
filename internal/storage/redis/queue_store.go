package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type QueueStore struct {
	client *Client
}

func NewQueueStore(client *Client) *QueueStore {
	return &QueueStore{client: client}
}

func (q *QueueStore) Enqueue(ctx context.Context, queue string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal queue payload: %w", err)
	}
	return q.client.GetDB().LPush(ctx, queue, payload).Err()
}

func (q *QueueStore) Dequeue(ctx context.Context, queue string) ([]byte, error) {
	result, err := q.client.GetDB().RPop(ctx, queue).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return result, err
}

func (q *QueueStore) Length(ctx context.Context, queue string) (int64, error) {
	return q.client.GetDB().LLen(ctx, queue).Result()
}