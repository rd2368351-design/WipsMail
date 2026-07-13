package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email shared.EmailAddress) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]User, int64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}