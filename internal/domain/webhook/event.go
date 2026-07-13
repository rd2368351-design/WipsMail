package webhook

type EventType string

const (
	EventSent         EventType = "message.sent"
	EventDelivered    EventType = "message.delivered"
	EventOpened       EventType = "message.opened"
	EventClicked      EventType = "message.clicked"
	EventBounced      EventType = "message.bounced"
	EventComplaint    EventType = "message.complaint"
	EventUnsubscribed EventType = "message.unsubscribed"
	EventFailed       EventType = "message.failed"
)

type Event struct {
	Type      EventType
	MessageID string
	TenantID  string
	Timestamp int64
	Data      map[string]any
}