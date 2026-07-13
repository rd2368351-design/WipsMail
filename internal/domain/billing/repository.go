package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Repository interface {
	SaveInvoice(ctx context.Context, invoice *Invoice) error
	FindInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error)
	ListInvoices(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]Invoice, int64, error)
	UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status InvoiceStatus) error
	SaveUsage(ctx context.Context, usage *UsageRecord) error
	GetUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) ([]UsageRecord, error)
	GetCurrentUsage(ctx context.Context, tenantID uuid.UUID) (*UsageRecord, error)
}