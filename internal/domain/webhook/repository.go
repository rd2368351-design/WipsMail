package webhook

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, webhook *Webhook) error
	FindByID(ctx context.Context, id uuid.UUID) (*Webhook, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]Webhook, int64, error)
	ListByEvent(ctx context.Context, tenantID uuid.UUID, event EventType) ([]Webhook, error)
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}