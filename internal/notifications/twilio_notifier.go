package notifications

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"strings"
)

// NewTwilioNotifierFromEnv builds a Notifier using Twilio env vars.
// Required:
//   - TWILIO_ACCOUNT_SID
//   - TWILIO_AUTH_TOKEN
//   - TWILIO_MESSAGING_SERVICE_SID
func NewTwilioNotifierFromEnv() (Notifier, error) {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	msgSvcSID := os.Getenv("TWILIO_MESSAGING_SERVICE_SID")

	if accountSID == "" {
		return nil, errors.New("TWILIO_ACCOUNT_SID not set")
	}
	if authToken == "" {
		return nil, errors.New("TWILIO_AUTH_TOKEN not set")
	}
	if msgSvcSID == "" {
		return nil, errors.New("TWILIO_MESSAGING_SERVICE_SID not set")
	}

	return &TwilioNotifier{
		accountSID:          accountSID,
		authToken:           authToken,
		messagingServiceSID: msgSvcSID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// TwilioNotifier implements Notifier using Twilio SMS.
type TwilioNotifier struct {
	accountSID          string
	authToken           string
	messagingServiceSID string
	httpClient          *http.Client
}

// Send sends SMS via Twilio. Only ChannelSMS is supported.
func (t *TwilioNotifier) Send(ctx context.Context, req Request) error {
	if req.Channel != ChannelSMS {
		return fmt.Errorf("twilio notifier only supports SMS channel, got %s", req.Channel)
	}
	if req.To == "" {
		return errors.New("missing To")
	}
	if req.Body == "" {
		return errors.New("missing Body")
	}

	form := url.Values{}
	form.Set("To", req.To)
	form.Set("MessagingServiceSid", t.messagingServiceSID)
	form.Set("Body", req.Body)

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, io.NopCloser(strings.NewReader(form.Encode())))
	if err != nil {
		return fmt.Errorf("build twilio request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(t.accountSID, t.authToken)

	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("twilio http error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("twilio send failed status=%d body=%s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("twilio send failed with status %d", resp.StatusCode)
	}

	log.Printf("twilio SMS sent to=%s len=%d", req.To, len(req.Body))
	return nil
}
