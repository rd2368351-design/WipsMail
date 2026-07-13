package message

import "github.com/wispmail/wispmail/internal/domain/shared"

type Sender struct {
	Address    shared.EmailAddress
	Domain     string
	Verified   bool
	SpfPassed  bool
	DkimPassed bool
}

func NewSender(address shared.EmailAddress) Sender {
	return Sender{
		Address: address,
		Domain:  address.Domain(),
	}
}