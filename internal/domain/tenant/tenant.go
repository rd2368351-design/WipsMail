package tenant

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	Active    bool
	Plan      string
	Settings  Settings
	Limits    Limits
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTenant(name, slug, plan string) *Tenant {
	now := time.Now().UTC()
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Active:    true,
		Plan:      plan,
		Settings:  DefaultSettings(),
		Limits:    DefaultLimits(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Tenant) Deactivate() {
	t.Active = false
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tenant) Activate() {
	t.Active = true
	t.UpdatedAt = time.Now().UTC()
}