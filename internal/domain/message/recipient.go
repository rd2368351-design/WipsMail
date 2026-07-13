package message

import "github.com/wispmail/wispmail/internal/domain/shared"

type RecipientType string

const (
	RecipientTo  RecipientType = "to"
	RecipientCc  RecipientType = "cc"
	RecipientBcc RecipientType = "bcc"
)

type Recipient struct {
	Address shared.EmailAddress
	Type    RecipientType
}

func NewRecipient(address shared.EmailAddress, rType RecipientType) Recipient {
	return Recipient{Address: address, Type: rType}
}