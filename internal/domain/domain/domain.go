package domain

import (
	"time"

	"github.com/google/uuid"
)

type Domain struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	Verified       bool
	DkimVerified   bool
	SpfVerified    bool
	DmarcVerified  bool
	Verification   Verification
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewDomain(tenantID uuid.UUID, name string) *Domain {
	now := time.Now().UTC()
	return &Domain{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         name,
		Verification: NewVerification(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (d *Domain) MarkVerified() {
	d.Verified = true
	d.UpdatedAt = time.Now().UTC()
}

func (d *Domain) IsReady() bool {
	return d.Verified && d.DkimVerified && d.SpfVerified
}