package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"dose-dock-tts-engine/internal/tts"
)

type ttsRequest struct {
	Text         string  `json:"text"`
	Prompt       string  `json:"prompt"`
	SpeakingRate float32 `json:"speakingRate"`
}

type ttsResponse struct {
	File        string `json:"file"`
	AudioBase64 string `json:"audioBase64"`
}

func (s *Server) handleTTSSpeak(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.tts == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("tts not configured"))
		return
	}

	var body ttsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	resp, err := s.tts.Synthesize(ctx, tts.SynthesizeRequest{
		Text:         body.Text,
		Prompt:       body.Prompt,
		SpeakingRate: body.SpeakingRate,
	})
	if err != nil {
		log.Printf("TTS synth error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("tts error"))
		return
	}

	filePath, err := saveMP3(resp.Audio)
	if err != nil {
		log.Printf("save MP3 error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to save audio"))
		return
	}

	if err := openWithDefaultPlayer(filePath); err != nil {
		log.Printf("auto play error: %v", err)
	}

	out := ttsResponse{
		File:        filePath,
		AudioBase64: base64.StdEncoding.EncodeToString(resp.Audio),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Printf("encode tts response error: %v", err)
	}
}

func saveMP3(data []byte) (string, error) {
	dir := "tts_output"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("tts-%d.mp3", time.Now().UnixMilli())
	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func openWithDefaultPlayer(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
