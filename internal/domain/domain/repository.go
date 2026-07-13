package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, domain *Domain) error
	FindByID(ctx context.Context, id uuid.UUID) (*Domain, error)
	FindByName(ctx context.Context, name string) (*Domain, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]Domain, int64, error)
	Update(ctx context.Context, domain *Domain) error
	Delete(ctx context.Context, id uuid.UUID) error
}