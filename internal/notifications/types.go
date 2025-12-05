package notifications

import (
	"context"
	"time"
)

type Channel string

const (
	ChannelSMS Channel = "SMS"
	ChannelTTS Channel = "TTS"
)

type Request struct {
	ID           string            // optional, for tracing
	To           string            // phone number in E.164
	Body         string            // plain text body
	Channel      Channel           // SMS or TTS
	ScheduledFor time.Time         // now or future
	Metadata     map[string]string // extra info if needed
}

type Notifier interface {
	Send(ctx context.Context, req Request) error
}
