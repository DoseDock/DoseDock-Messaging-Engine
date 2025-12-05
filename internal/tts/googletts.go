package tts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient  *http.Client
	accessToken string
	projectID   string
}

func NewClientFromEnv() (*Client, error) {
	token := os.Getenv("GOOGLE_TTS_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GOOGLE_TTS_ACCESS_TOKEN not set")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT not set")
	}

	return &Client{
		httpClient:  &http.Client{Timeout: 20 * time.Second},
		accessToken: token,
		projectID:   projectID,
	}, nil
}

// Synthesize implements TTSClient and returns raw MP3 bytes.
func (c *Client) Synthesize(ctx context.Context, req SynthesizeRequest) (*SynthesizeResponse, error) {
	body := map[string]any{
		"input": map[string]string{
			"text": req.Text,
		},
		"voice": map[string]string{
			"languageCode": "en-US",
			"name":         "en-US-Standard-C",
		},
		"audioConfig": map[string]any{
			"audioEncoding": "MP3",
			// you can add speaking rate here if you want:
			// "speakingRate": req.SpeakingRate,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode tts request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://texttospeech.googleapis.com/v1/text:synthesize",
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("build tts request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-user-project", c.projectID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tts http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts non 200: %d, body=%s", resp.StatusCode, string(b))
	}

	var respBody struct {
		AudioContent string `json:"audioContent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("decode tts response: %w", err)
	}
	if respBody.AudioContent == "" {
		return nil, fmt.Errorf("empty audioContent in tts response")
	}

	audioBytes, err := base64.StdEncoding.DecodeString(respBody.AudioContent)
	if err != nil {
		return nil, fmt.Errorf("decode base64 audioContent: %w", err)
	}

	return &SynthesizeResponse{Audio: audioBytes}, nil
}
