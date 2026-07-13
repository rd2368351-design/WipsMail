package shared

import "fmt"

type Operator string

const (
	OpEquals      Operator = "eq"
	OpNotEquals   Operator = "neq"
	OpContains    Operator = "contains"
	OpGreaterThan Operator = "gt"
	OpLessThan    Operator = "lt"
	OpIn          Operator = "in"
	OpBetween     Operator = "between"
	OpStartsWith  Operator = "startswith"
)

type Filter struct {
	Field    string
	Operator Operator
	Value    any
	Values   []any
}

func NewFilter(field string, op Operator, value any) Filter {
	return Filter{
		Field:    sanitizeIdentifier(field),
		Operator: op,
		Value:    value,
	}
}

func NewInFilter(field string, values []any) Filter {
	return Filter{
		Field:    sanitizeIdentifier(field),
		Operator: OpIn,
		Values:   values,
	}
}

type FilterSet struct {
	Filters []Filter
}

func NewFilterSet() *FilterSet {
	return &FilterSet{Filters: make([]Filter, 0)}
}

func (fs *FilterSet) Add(filter Filter) *FilterSet {
	fs.Filters = append(fs.Filters, filter)
	return fs
}

func (fs *FilterSet) IsEmpty() bool {
	return len(fs.Filters) == 0
}

func (fs *FilterSet) String() string {
	return fmt.Sprintf("FilterSet{filters: %v}", fs.Filters)
}