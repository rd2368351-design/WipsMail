package shared

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrAlreadyExists       ErrorCode = "ALREADY_EXISTS"
	ErrInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrConflict            ErrorCode = "CONFLICT"
	ErrInternalError       ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrRateLimited         ErrorCode = "RATE_LIMITED"
	ErrQuotaExceeded       ErrorCode = "QUOTA_EXCEEDED"
	ErrExpired             ErrorCode = "EXPIRED"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewError(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func WrapError(code ErrorCode, message string, err error) *DomainError {
	return &DomainError{Code: code, Message: message, Err: err}
}

var (
	ErrNotFoundError       = errors.New("resource not found")
	ErrAlreadyExistsError  = errors.New("resource already exists")
	ErrInvalidInputError   = errors.New("invalid input")
	ErrUnauthorizedError   = errors.New("unauthorized")
	ErrForbiddenError      = errors.New("forbidden")
)