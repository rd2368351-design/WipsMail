package tenant

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, tenant *Tenant) error
	FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*Tenant, error)
	List(ctx context.Context, pagination shared.Pagination) ([]Tenant, int64, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}