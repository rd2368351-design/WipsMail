package billing

import (
	"time"

	"github.com/google/uuid"
)

type UsageRecord struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Date          time.Time
	EmailsSent    int64
	EmailsFailed  int64
	EmailsBounced int64
	StorageBytes  int64
	APICalls      int64
}

func NewUsageRecord(tenantID uuid.UUID) *UsageRecord {
	return &UsageRecord{
		ID:       uuid.New(),
		TenantID: tenantID,
		Date:     time.Now().UTC().Truncate(24 * time.Hour),
	}
}

func (u *UsageRecord) AddEmail(status string) {
	switch status {
	case "sent", "delivered":
		u.EmailsSent++
	case "failed":
		u.EmailsFailed++
	case "bounced":
		u.EmailsBounced++
	}
}