package notifications

import (
	"fmt"
)

type EventType string

const (
	EventDoseDue         EventType = "DOSE_DUE"
	EventRefillReminder  EventType = "REFILL_REMINDER"
	EventTestReminder    EventType = "TEST_REMINDER"
)

type EventPayload struct {
	Event   EventType
	To      string
	Payload map[string]string // e.g. {"patientName": "...", "time": "..."}
}

// RenderBody builds a message body from the event type and payload
func RenderBody(ev EventPayload) (string, error) {
	switch ev.Event {
	case EventDoseDue:
		// expects: patientName, time, meds
		return fmt.Sprintf(
			"Hi %s, this is your DoseDock reminder to take your %s at %s.",
			ev.Payload["patientName"],
			ev.Payload["meds"],
			ev.Payload["time"],
		), nil

	case EventRefillReminder:
		// expects: patientName, meds
		return fmt.Sprintf(
			"Hi %s, your DoseDock dispenser is running low on %s. Please refill soon.",
			ev.Payload["patientName"],
			ev.Payload["meds"],
		), nil

	case EventTestReminder:
		return "DoseDock test reminder, notifications are working.", nil

	default:
		return "", fmt.Errorf("unknown event type %s", ev.Event)
	}
}
