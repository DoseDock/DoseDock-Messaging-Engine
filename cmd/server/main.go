package main

import (
	"log"
	"net/http"
	"os"

	"dose-dock-tts-engine/internal/httpapi"
	"dose-dock-tts-engine/internal/notifications"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	twilioNotifier, err := notifications.NewTwilioSMSNotifierFromEnv()
	if err != nil {
		log.Fatalf("failed to init Twilio notifier: %v", err)
	}

	srv := httpapi.NewServer(twilioNotifier)

	log.Printf("notification engine listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
