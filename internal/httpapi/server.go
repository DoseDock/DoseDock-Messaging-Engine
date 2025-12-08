package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"sync"
	"fmt"

	"dose-dock-tts-engine/internal/notifications"
	"dose-dock-tts-engine/internal/tts"
)

type Server struct {
	notifier notifications.Notifier
	tts      tts.TTSClient
	mux      *http.ServeMux
}

func NewServer(notifier notifications.Notifier, ttsClient tts.TTSClient) *Server {
	s := &Server{
		notifier: notifier,
		tts:      ttsClient,
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /send-sms", s.handleSendSMS)
	s.mux.HandleFunc("POST /send-event", s.handleSendEvent)
	s.mux.HandleFunc("POST /twilio/status", s.handleTwilioStatus)
	s.mux.HandleFunc("POST /tts/speak", s.handleTTSSpeak)

	// Serve the UI from /ui/
	fileServer := http.FileServer(http.Dir("web"))
	s.mux.Handle("/ui/", http.StripPrefix("/ui/", fileServer))

	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

// existing types and handlers you already had:

type sendSMSRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

type sendEventRequest struct {
	Event        string            `json:"event"`
	To           string            `json:"to"`
	Payload      map[string]string `json:"payload"`
	Voice        string            `json:"voice,omitempty"`
	Emotion      string            `json:"emotion,omitempty"`
	SpeakingRate float32           `json:"speakingRate,omitempty"`
	Prompt       string            `json:"prompt,omitempty"`
}

func (s *Server) handleSendSMS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body sendSMSRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req := notifications.Request{
		To:      body.To,
		Body:    body.Body,
		Channel: notifications.ChannelSMS,
	}

	if err := s.notifier.Send(ctx, req); err != nil {
		log.Printf("SendSMS error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("failed to send"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleSendEvent(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var body sendEventRequest
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        _, _ = w.Write([]byte("invalid json"))
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    ev := notifications.EventPayload{
        Event:        notifications.EventType(body.Event),
        To:           body.To,
        Payload:      body.Payload,
        Voice:        body.Voice,
        Emotion:      body.Emotion,
        SpeakingRate: body.SpeakingRate,
        Prompt:       body.Prompt,
    }

    log.Printf("handleSendEvent voice=%q emotion=%q rate=%v prompt=%q",
        body.Voice, body.Emotion, body.SpeakingRate, body.Prompt)

    text, err := notifications.RenderBody(ev)
    if err != nil {
        log.Printf("RenderBody error: %v", err)
        w.WriteHeader(http.StatusBadRequest)
        _, _ = w.Write([]byte("bad event or payload"))
        return
    }

    // Single path: send SMS + TTS inside this helper
    if err := s.sendSMSAndSpeak(ctx, ev, text); err != nil {
        log.Printf("sendSMSAndSpeak error: %v", err)
        w.WriteHeader(http.StatusBadGateway)
        _, _ = w.Write([]byte("failed to send or speak"))
        return
    }

    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleTwilioStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Twilio sends application/x-www-form-urlencoded
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	messageSid := r.FormValue("MessageSid")
	messageStatus := r.FormValue("MessageStatus")
	to := r.FormValue("To")
	from := r.FormValue("From")
	errorCode := r.FormValue("ErrorCode")

	log.Printf(
		"Twilio status callback: sid=%s status=%s to=%s from=%s errorCode=%s",
		messageSid, messageStatus, to, from, errorCode,
	)

	// Later you can persist this to a database instead of just logging
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) sendSMSAndSpeak(ctx context.Context, ev notifications.EventPayload, text string) error {
	smsReq := notifications.Request{
		To:      ev.To,
		Body:    text,
		Channel: notifications.ChannelSMS,
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// SMS
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.notifier.Send(ctx, smsReq); err != nil {
			errCh <- fmt.Errorf("sms send error: %w", err)
		}
	}()

	// TTS with caregiver prefs
	if s.tts != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			voice := ev.Voice
			if voice == "" {
				voice = "Charon"
			}

			rate := ev.SpeakingRate
			if rate <= 0 {
				rate = 1.0
			}

			prompt := ev.Prompt
			if prompt == "" {
				prompt = "Speak clearly and calmly for an older adult."
			}
			if ev.Emotion != "" {
				prompt = fmt.Sprintf("%s Use a %s tone.", prompt, ev.Emotion)
			}

			resp, err := s.tts.Synthesize(ctx, tts.SynthesizeRequest{
				Text:         text,
				Prompt:       prompt,
				SpeakingRate: rate,
				Voice:        voice,
				Emotion:      tts.Emotion(ev.Emotion),
			})
			if err != nil {
				errCh <- fmt.Errorf("tts synth error: %w", err)
				return
			}

			// Save and auto play, same as /tts/speak
			path, err := saveMP3(resp.Audio)
			if err != nil {
				errCh <- fmt.Errorf("save MP3 error: %w", err)
				return
			}

			if err := openWithDefaultPlayer(path); err != nil {
				// not fatal but log it
				log.Printf("auto play error: %v", err)
			}
		}()
	} else {
		log.Printf("sendSMSAndSpeak: tts client is nil, skipping voice playback")
	}

	// Wait for both goroutines, but respect context
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(errCh)
		if len(errCh) == 0 {
			return nil
		}
		return <-errCh
	case <-ctx.Done():
		return ctx.Err()
	}
}
