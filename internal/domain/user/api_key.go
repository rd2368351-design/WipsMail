package user

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TenantID    uuid.UUID
	Name        string
	KeyPrefix   string
	KeyHash     string
	Permissions []Permission
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	RevokedAt   *time.Time
}

func NewAPIKey(userID, tenantID uuid.UUID, name string, keyHash string, permissions []Permission) *APIKey {
	return &APIKey{
		ID:          uuid.New(),
		UserID:      userID,
		TenantID:    tenantID,
		Name:        name,
		KeyHash:     keyHash,
		KeyPrefix:   name[:8],
		Permissions: permissions,
		CreatedAt:   time.Now().UTC(),
	}
}

func (k *APIKey) Revoke() {
	now := time.Now().UTC()
	k.RevokedAt = &now
}

func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*k.ExpiresAt)
}

func (k *APIKey) IsValid() bool {
	return !k.IsRevoked() && !k.IsExpired()
}