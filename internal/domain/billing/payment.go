package billing

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID        uuid.UUID
	InvoiceID uuid.UUID
	Amount    float64
	Currency  string
	Method    string
	Status    string
	GatewayID string
	CreatedAt time.Time
}

func NewPayment(invoiceID uuid.UUID, amount float64, currency, method string) *Payment {
	return &Payment{
		ID:        uuid.New(),
		InvoiceID: invoiceID,
		Amount:    amount,
		Currency:  currency,
		Method:    method,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}
}