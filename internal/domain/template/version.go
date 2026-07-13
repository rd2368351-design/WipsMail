package template

import (
	"time"

	"github.com/google/uuid"
)

type Version struct {
	ID         uuid.UUID
	TemplateID uuid.UUID
	Version    int
	Subject    string
	HTMLBody   string
	PlainBody  string
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
}

func NewVersion(templateID, createdBy uuid.UUID, version int, subject, htmlBody, plainBody string) *Version {
	return &Version{
		ID:         uuid.New(),
		TemplateID: templateID,
		Version:    version,
		Subject:    subject,
		HTMLBody:   htmlBody,
		PlainBody:  plainBody,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now().UTC(),
	}
}