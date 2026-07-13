package template

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Subject   string
	HTMLBody  string
	PlainBody string
	Version   int
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTemplate(tenantID uuid.UUID, name, subject, htmlBody, plainBody string) *Template {
	now := time.Now().UTC()
	return &Template{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		Subject:   subject,
		HTMLBody:  htmlBody,
		PlainBody: plainBody,
		Version:   1,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}