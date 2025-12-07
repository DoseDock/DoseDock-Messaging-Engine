package tts

import "context"

type Emotion string

const (
	EmotionNeutral  Emotion = "neutral"
	EmotionCalm     Emotion = "calm"
	EmotionFriendly Emotion = "friendly"
	EmotionUrgent   Emotion = "urgent"
)

type SynthesizeRequest struct {
	Text         string
	Prompt       string
	SpeakingRate float32
	Voice        string
	Emotion      Emotion
}

type SynthesizeResponse struct {
	Audio []byte
}

type TTSClient interface {
	Synthesize(ctx context.Context, req SynthesizeRequest) (*SynthesizeResponse, error)
}
