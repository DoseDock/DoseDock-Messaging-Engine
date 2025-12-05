package tts

import "context"

type SynthesizeRequest struct {
	Text         string
	Prompt       string
	SpeakingRate float32
}

type SynthesizeResponse struct {
	Audio []byte // raw MP3 bytes
}

type TTSClient interface {
	Synthesize(ctx context.Context, req SynthesizeRequest) (*SynthesizeResponse, error)
}
