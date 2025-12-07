package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"dose-dock-tts-engine/internal/httpapi"
	"dose-dock-tts-engine/internal/notifications"
	"dose-dock-tts-engine/internal/tts"
)

func main() {
	// Load env vars from .env.local if present
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("no .env.local file loaded: %v", err)
	}

	notifier, err := notifications.NewTwilioSMSNotifierFromEnv()
	if err != nil {
		log.Fatalf("twilio notifier init: %v", err)
	}

	ttsClient, err := tts.NewClientFromEnv()
	if err != nil {
		log.Fatalf("tts client init: %v", err)
	}

	srv := httpapi.NewServer(notifier, ttsClient)

	log.Println("listening on :8090")
	log.Fatal(http.ListenAndServe(":8090", srv))
}