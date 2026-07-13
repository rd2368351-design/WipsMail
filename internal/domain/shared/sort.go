package shared

import (
	"fmt"
	"strings"
)

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type Sort struct {
	Field string
	Order SortOrder
}

func NewSort(field string, order SortOrder) Sort {
	normalizedOrder := SortOrder(strings.ToUpper(strings.TrimSpace(string(order))))
	if normalizedOrder != SortAsc && normalizedOrder != SortDesc {
		normalizedOrder = SortAsc
	}
	return Sort{
		Field: sanitizeIdentifier(field),
		Order: normalizedOrder,
	}
}

func (s Sort) String() string {
	if s.Field == "" {
		return "created_at DESC"
	}
	return fmt.Sprintf("%s %s", s.Field, s.Order)
}

func sanitizeIdentifier(field string) string {
	trimmed := strings.TrimSpace(field)
	if trimmed == "" {
		return "created_at"
	}
	var builder strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}
	result := builder.String()
	if result == "" {
		return "created_at"
	}
	return result
}