package webhook

import (
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	URL       string
	Secret    string
	Events    []string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewWebhook(tenantID uuid.UUID, name, url, secret string, events []string) *Webhook {
	now := time.Now().UTC()
	return &Webhook{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}