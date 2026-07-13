package audit

import (
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionLogin        Action = "user.login"
	ActionLogout       Action = "user.logout"
	ActionSendEmail    Action = "email.send"
	ActionDeleteEmail  Action = "email.delete"
	ActionCreateUser   Action = "user.create"
	ActionDeleteUser   Action = "user.delete"
	ActionUpdateTenant Action = "tenant.update"
	ActionCreateAPIKey Action = "apikey.create"
	ActionRevokeAPIKey Action = "apikey.revoke"
)

type Event struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	ActorID      uuid.UUID
	ActorEmail   string
	Action       Action
	Resource     string
	ResourceID   string
	Details      map[string]any
	IPAddress    string
	UserAgent    string
	Success      bool
	ErrorMessage string
	CreatedAt    time.Time
}

func NewEvent(tenantID, actorID uuid.UUID, actorEmail string, action Action, resource, resourceID string) *Event {
	return &Event{
		ID:         uuid.New(),
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorEmail: actorEmail,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Success:    true,
		CreatedAt:  time.Now().UTC(),
	}
}

func (e *Event) MarkFailed(errMsg string) {
	e.Success = false
	e.ErrorMessage = errMsg
}