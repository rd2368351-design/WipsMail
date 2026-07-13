package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type User struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Email         shared.EmailAddress
	Username      string
	PasswordHash  string
	FirstName     string
	LastName      string
	Role          Role
	EmailVerified bool
	MFAEnabled    bool
	Active        bool
	LastLoginAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewUser(tenantID uuid.UUID, email shared.EmailAddress, username, passwordHash string) *User {
	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now().UTC()
}