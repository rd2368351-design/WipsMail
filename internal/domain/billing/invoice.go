package billing

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceDraft   InvoiceStatus = "draft"
	InvoiceOpen    InvoiceStatus = "open"
	InvoicePaid    InvoiceStatus = "paid"
	InvoiceVoid    InvoiceStatus = "void"
	InvoiceOverdue InvoiceStatus = "overdue"
)

type Invoice struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	InvoiceNumber string
	Status        InvoiceStatus
	Amount        float64
	Currency      string
	Description   string
	PeriodStart   time.Time
	PeriodEnd     time.Time
	DueDate       time.Time
	PaidAt        *time.Time
	CreatedAt     time.Time
}

func NewInvoice(tenantID uuid.UUID, amount float64, currency string, periodStart, periodEnd time.Time) *Invoice {
	return &Invoice{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Status:      InvoiceDraft,
		Amount:      amount,
		Currency:    currency,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		DueDate:     time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt:   time.Now().UTC(),
	}
}