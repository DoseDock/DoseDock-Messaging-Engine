package notifications

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"log"
)

type TwilioSMSNotifier struct {
	accountSID          string
	authToken           string
	messagingServiceSID string
	httpClient          *http.Client

	maxPerSecond int
	lastSend     time.Time
}

func NewTwilioSMSNotifierFromEnv() (*TwilioSMSNotifier, error) {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	messagingServiceSID := os.Getenv("TWILIO_MESSAGING_SERVICE_SID")

	if accountSID == "" || authToken == "" || messagingServiceSID == "" {
		return nil, errors.New("TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_MESSAGING_SERVICE_SID must be set")
	}

		return &TwilioSMSNotifier{
		accountSID:          accountSID,
		authToken:           authToken,
		messagingServiceSID: messagingServiceSID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxPerSecond: 5, // simple soft limit
	}, nil

}

func (n *TwilioSMSNotifier) Send(ctx context.Context, req Request) error {
	if req.Channel != ChannelSMS {
		return nil
	}
	if req.To == "" {
		return errors.New("missing To")
	}
	if req.Body == "" {
		return errors.New("missing Body")
	}

	// simple rate limiting: ensure at most maxPerSecond
	now := time.Now()
	if !n.lastSend.IsZero() {
		minInterval := time.Second / time.Duration(n.maxPerSecond)
		if since := now.Sub(n.lastSend); since < minInterval {
			sleep := minInterval - since
			time.Sleep(sleep)
		}
	}
	n.lastSend = time.Now()

	endpoint := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		n.accountSID,
	)

	form := url.Values{}
	form.Set("To", req.To)
	form.Set("MessagingServiceSid", n.messagingServiceSID)
	form.Set("Body", req.Body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build twilio request: %w", err)
	}

	httpReq.SetBasicAuth(n.accountSID, n.authToken)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	var bodyBytes []byte

	// simple retry loop
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err = n.httpClient.Do(httpReq)
		if err != nil {
			if attempt == maxAttempts {
				return fmt.Errorf("twilio request failed after retries: %w", err)
			}
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		bodyBytes, _ = io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// success
			log.Printf("Twilio SMS sent to %s, status=%d", req.To, resp.StatusCode)
			return nil
		}

		// retry on 429 and 5xx
		if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
			log.Printf("Twilio temp error, attempt %d, status=%d body=%s", attempt, resp.StatusCode, string(bodyBytes))
			if attempt == maxAttempts {
				return fmt.Errorf("twilio error after retries: status=%d body=%s", resp.StatusCode, string(bodyBytes))
			}
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		// non retryable error
		log.Printf("Twilio permanent error, status=%d body=%s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("twilio error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
