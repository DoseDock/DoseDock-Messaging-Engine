package notifications

import "fmt"

type EventType string

const (
	EventDoseDue        EventType = "DOSE_DUE"
	EventRefillReminder EventType = "REFILL_REMINDER"
	EventTestReminder   EventType = "TEST_REMINDER"
)

type EventPayload struct {
	Event   EventType         `json:"event"`
	To      string            `json:"to"`
	Payload map[string]string `json:"payload"` // e.g. {"patientName": "...", "time": "..."}

	// TTS preferences coming from caregiver UI
	Voice        string  `json:"voice,omitempty"`        // "Charon", "Kore", etc
	Emotion      string  `json:"emotion,omitempty"`      // "calm", "friendly", "urgent"
	SpeakingRate float32 `json:"speakingRate,omitempty"` // 0.8, 1.0, etc
	Prompt       string  `json:"prompt,omitempty"`       // extra style instructions
}

// RenderBody builds a message body from the event type and payload
func RenderBody(ev EventPayload) (string, error) {
	switch ev.Event {
	case EventDoseDue:
		return fmt.Sprintf(
			"Hi %s, this is your DoseDock reminder to take your %s at %s.",
			ev.Payload["patientName"],
			ev.Payload["meds"],
			ev.Payload["time"],
		), nil

	case EventRefillReminder:
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
