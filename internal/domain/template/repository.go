package template

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, template *Template) error
	FindByID(ctx context.Context, id uuid.UUID) (*Template, error)
	FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*Template, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]Template, int64, error)
	Update(ctx context.Context, template *Template) error
	Delete(ctx context.Context, id uuid.UUID) error
	SaveVersion(ctx context.Context, version *Version) error
	ListVersions(ctx context.Context, templateID uuid.UUID) ([]Version, error)
}