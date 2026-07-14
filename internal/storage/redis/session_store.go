package redis

import (
	"context"
	"fmt"
	"time"
)

type SessionStore struct {
	client *Client
}

func NewSessionStore(client *Client) *SessionStore {
	return &SessionStore{client: client}
}

func (s *SessionStore) Save(ctx context.Context, sessionID string, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return s.client.GetDB().Set(ctx, key, data, ttl).Err()
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) ([]byte, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	return s.client.GetDB().Get(ctx, key).Bytes()
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return s.client.GetDB().Del(ctx, key).Err()
}