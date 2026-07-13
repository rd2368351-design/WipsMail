package shared

import (
	"fmt"
	"net/mail"
	"strings"
)

type EmailAddress struct {
	value string
}

func NewEmailAddress(address string) (EmailAddress, error) {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return EmailAddress{}, NewError(ErrInvalidInput, "email address cannot be empty")
	}
	if len(trimmed) > 320 {
		return EmailAddress{}, NewError(ErrInvalidInput, "email address exceeds maximum length of 320 characters")
	}
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return EmailAddress{}, WrapError(ErrInvalidInput, "invalid email address format", err)
	}
	if !strings.Contains(parsed.Address, "@") {
		return EmailAddress{}, NewError(ErrInvalidInput, "email address must contain @ symbol")
	}
	parts := strings.SplitN(parsed.Address, "@", 2)
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return EmailAddress{}, NewError(ErrInvalidInput, "email address must have local and domain parts")
	}
	return EmailAddress{value: parsed.Address}, nil
}

func MustNewEmailAddress(address string) EmailAddress {
	addr, err := NewEmailAddress(address)
	if err != nil {
		panic(fmt.Sprintf("invalid email address %q: %v", address, err))
	}
	return addr
}

func (e EmailAddress) String() string {
	return e.value
}

func (e EmailAddress) Domain() string {
	parts := strings.SplitN(e.value, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

func (e EmailAddress) LocalPart() string {
	parts := strings.SplitN(e.value, "@", 2)
	if len(parts) != 2 {
		return e.value
	}
	return parts[0]
}

func (e EmailAddress) Equal(other EmailAddress) bool {
	return strings.EqualFold(e.value, other.value)
}

func (e EmailAddress) IsEmpty() bool {
	return e.value == ""
}

func (e EmailAddress) MarshalText() ([]byte, error) {
	return []byte(e.value), nil
}

func (e EmailAddress) UnmarshalText(text []byte) error {
	addr, err := NewEmailAddress(string(text))
	if err != nil {
		return err
	}
	e.value = addr.value
	return nil
}