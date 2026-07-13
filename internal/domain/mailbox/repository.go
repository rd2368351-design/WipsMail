package mailbox

import (
	"context"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	Save(ctx context.Context, mailbox *Mailbox) error
	FindByID(ctx context.Context, id uuid.UUID) (*Mailbox, error)
	FindByEmail(ctx context.Context, email shared.EmailAddress) (*Mailbox, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Mailbox, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]Mailbox, int64, error)
	Update(ctx context.Context, mailbox *Mailbox) error
	Delete(ctx context.Context, id uuid.UUID) error
}