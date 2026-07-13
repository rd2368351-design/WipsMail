package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Verification struct {
	Token        string
	RecordType   DNSRecordType
	RecordHost   string
	RecordValue  string
	ExpiresAt    time.Time
}

func NewVerification() Verification {
	token := generateToken(32)
	return Verification{
		Token:       token,
		RecordType:  RecordTXT,
		RecordHost:  "@",
		RecordValue: "wispmail-verify=" + token,
		ExpiresAt:   time.Now().UTC().Add(72 * time.Hour),
	}
}

func (v Verification) IsExpired() bool {
	return time.Now().UTC().After(v.ExpiresAt)
}

func generateToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}