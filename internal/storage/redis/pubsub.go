package redis

import (
	"context"
	"fmt"
)

type PubSub struct {
	client *Client
}

func NewPubSub(client *Client) *PubSub {
	return &PubSub{client: client}
}

func (p *PubSub) Publish(ctx context.Context, channel string, message []byte) error {
	return p.client.GetDB().Publish(ctx, channel, message).Err()
}

func (p *PubSub) Subscribe(ctx context.Context, channel string, handler func([]byte) error) error {
	sub := p.client.GetDB().Subscribe(ctx, channel)
	defer sub.Close()

	ch := sub.Channel()

	for {
		select {
		case msg := <-ch:
			if err := handler([]byte(msg.Payload)); err != nil {
				return fmt.Errorf("handler error: %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}