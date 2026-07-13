package message

import (
	"time"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Message struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	MessageID   string
	InReplyTo   string
	References  string
	Subject     string
	From        shared.EmailAddress
	Sender      shared.EmailAddress
	ReplyTo     shared.EmailAddress
	To          []shared.EmailAddress
	Cc          []shared.EmailAddress
	Bcc         []shared.EmailAddress
	Body        Body
	Attachments []Attachment
	Headers     []Header
	Status      Status
	Priority    Priority
	Size        int64
	ReceivedAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewMessage(tenantID uuid.UUID, from shared.EmailAddress, to []shared.EmailAddress, subject string) *Message {
	now := time.Now().UTC()
	return &Message{
		ID:         uuid.New(),
		TenantID:   tenantID,
		From:       from,
		To:         to,
		Subject:    subject,
		Status:     StatusPending,
		Priority:   PriorityNormal,
		ReceivedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (m *Message) MarkProcessing() {
	m.Status = StatusProcessing
	m.UpdatedAt = time.Now().UTC()
}

func (m *Message) MarkDelivered() {
	m.Status = StatusDelivered
	m.UpdatedAt = time.Now().UTC()
}

func (m *Message) MarkFailed() {
	m.Status = StatusFailed
	m.UpdatedAt = time.Now().UTC()
}

func (m *Message) MarkBounced() {
	m.Status = StatusBounced
	m.UpdatedAt = time.Now().UTC()
}