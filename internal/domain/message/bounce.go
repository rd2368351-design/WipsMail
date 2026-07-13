package message

import (
	"time"

	"github.com/google/uuid"
)

type BounceType string

const (
	BounceHard BounceType = "hard"
	BounceSoft BounceType = "soft"
)

type Bounce struct {
	ID          uuid.UUID
	MessageID   uuid.UUID
	Type        BounceType
	Code        string
	Description string
	BouncedAt   time.Time
}

func NewBounce(messageID uuid.UUID, bounceType BounceType, code, description string) *Bounce {
	return &Bounce{
		ID:          uuid.New(),
		MessageID:   messageID,
		Type:        bounceType,
		Code:        code,
		Description: description,
		BouncedAt:   time.Now().UTC(),
	}
}