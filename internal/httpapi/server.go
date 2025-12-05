package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	Event   string            `json:"event"`
	To      string            `json:"to"`
	Payload map[string]string `json:"payload"`
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
		Event:   notifications.EventType(body.Event),
		To:      body.To,
		Payload: body.Payload,
	}

	text, err := notifications.RenderBody(ev)
	if err != nil {
		log.Printf("RenderBody error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad event or payload"))
		return
	}

	req := notifications.Request{
		To:      ev.To,
		Body:    text,
		Channel: notifications.ChannelSMS,
	}

	if err := s.notifier.Send(ctx, req); err != nil {
		log.Printf("Send event SMS error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("failed to send"))
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
