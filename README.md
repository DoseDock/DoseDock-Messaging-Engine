# DoseDock Messaging Engine

DoseDock Notify is a lightweight messaging and voice engine that powers SMS reminders and spoken pillbox announcements for the DoseDock medication-adherence system.  
It provides a simple HTTP API plus a local testing UI for caregivers to preview voices, verify patient phone numbers, and send test reminders.

---

## Features

- **SMS Delivery (Twilio Messaging Service)**
  - Fully templated reminder generation
  - Status callbacks supported
- **Voice Synthesis (Google TTS, Chirp-3, Charon voice)**
  - Configurable voice, emotion, speaking rate, and prompt style
  - Local audio playback for testing
- **Caregiver UI**
  - Live preview of “what the patient hears”
  - Phone-number verification workflow with displayed Twilio validation code
  - Play-test voice, send sample reminders, and view history log
- **Unified Reminder Engine**
  - Single `/send-event` endpoint handles SMS + TTS simultaneously
  - Supports multiple event types: dose due, refill, test

## Environment Variables

Create a `.env` file:
```
TWILIO_ACCOUNT_SID=ACxxxx
TWILIO_AUTH_TOKEN=xxxx
TWILIO_MESSAGING_SERVICE_SID=MGxxxx
GOOGLE_APPLICATION_CREDENTIALS=./gcloud-creds.json
```

Your Google ADC credentials are stored automatically by:

`gcloud auth application-default login`


## Running the Server

```
go run ./cmd/server
```
Server runs at:
```
http://localhost:8090
```
Caregiver UI:

```
http://localhost:8090/ui/
```
## API Overview

To send a `POST /send-event` and trigger an SMS + TTS reminder.

```
{
  "event": "DOSE_DUE",
  "to": "+15551230000",
  "payload": {
    "patientName": "Jinal",
    "time": "8:00 pm",
    "meds": "evening medications"
  },
  "voice": "Charon",
  "emotion": "calm",
  "speakingRate": 1.0,
  "prompt": "Speak clearly for an older adult."
}
```

## Next Steps
- Randomized phrasing to avoid alert fatigue
    - either use gpt to randomize or have set list of messages to choose from
