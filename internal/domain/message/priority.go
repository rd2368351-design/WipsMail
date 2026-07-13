package message

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

func (p Priority) Weight() int {
	switch p {
	case PriorityLow:
		return 1
	case PriorityNormal:
		return 2
	case PriorityHigh:
		return 3
	case PriorityUrgent:
		return 4
	}
	return 2
}