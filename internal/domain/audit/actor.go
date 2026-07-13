package audit

import "github.com/google/uuid"

type Actor struct {
	ID    uuid.UUID
	Email string
	Role  string
}

func NewActor(id uuid.UUID, email, role string) Actor {
	return Actor{ID: id, Email: email, Role: role}
}

func SystemActor() Actor {
	return Actor{
		ID:    uuid.Nil,
		Email: "system@wispmail.internal",
		Role:  "system",
	}
}