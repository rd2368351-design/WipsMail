package message

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusQueued     Status = "queued"
	StatusDelivered  Status = "delivered"
	StatusFailed     Status = "failed"
	StatusBounced    Status = "bounced"
	StatusDeferred   Status = "deferred"
	StatusSpam       Status = "spam"
	StatusRejected   Status = "rejected"
)

func (s Status) IsFinal() bool {
	return s == StatusDelivered || s == StatusFailed || s == StatusBounced || s == StatusRejected
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusQueued, StatusDelivered,
		StatusFailed, StatusBounced, StatusDeferred, StatusSpam, StatusRejected:
		return true
	}
	return false
}