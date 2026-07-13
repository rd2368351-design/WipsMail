package webhook

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryPending  DeliveryStatus = "pending"
	DeliverySuccess  DeliveryStatus = "success"
	DeliveryFailed   DeliveryStatus = "failed"
	DeliveryRetrying DeliveryStatus = "retrying"
)

type Delivery struct {
	ID            uuid.UUID
	WebhookID     uuid.UUID
	Event         Event
	Status        DeliveryStatus
	StatusCode    int
	ResponseBody  string
	AttemptCount  int
	LastAttemptAt *time.Time
	NextAttemptAt *time.Time
	DeliveredAt   *time.Time
	CreatedAt     time.Time
}

func NewDelivery(webhookID uuid.UUID, event Event) *Delivery {
	return &Delivery{
		ID:        uuid.New(),
		WebhookID: webhookID,
		Event:     event,
		Status:    DeliveryPending,
		CreatedAt: time.Now().UTC(),
	}
}