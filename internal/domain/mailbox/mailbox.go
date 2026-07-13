package mailbox

import (
	"time"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type Mailbox struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	UserID       uuid.UUID
	Email        shared.EmailAddress
	Name         string
	Folders      []Folder
	Quota        Quota
	MessageCount int64
	UnreadCount  int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewMailbox(tenantID, userID uuid.UUID, email shared.EmailAddress, name string) *Mailbox {
	now := time.Now().UTC()
	return &Mailbox{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Email:     email,
		Name:      name,
		Quota:     DefaultQuota(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (m *Mailbox) IsOverQuota() bool {
	return m.MessageCount >= m.Quota.MaxMessages
}