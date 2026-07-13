package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, msg *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	FindByMessageID(ctx context.Context, messageID string) (*Message, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination, filters *shared.FilterSet, sort shared.Sort) ([]Message, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	Delete(ctx context.Context, id uuid.UUID) error
}