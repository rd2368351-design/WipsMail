package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, event *Event) error
	FindByID(ctx context.Context, id uuid.UUID) (*Event, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]Event, int64, error)
	ListByActor(ctx context.Context, actorID uuid.UUID, pagination shared.Pagination) ([]Event, int64, error)
	ListByAction(ctx context.Context, tenantID uuid.UUID, action Action, pagination shared.Pagination) ([]Event, int64, error)
}