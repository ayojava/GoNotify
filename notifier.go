package main

import (
	"encoding/json"
	"fmt"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// https://github.com/twilio/twilio-go
type Notifier interface {
	Send(to, body string) error
}

// TwilioNotifier implements Notifier using the official Twilio Go SDK.
type TwilioNotifier struct {
	FromNumber string // e.g. "whatsapp:+14155238886"
	client     *twilio.RestClient
}

func NewTwilioNotifier(accountSID, authToken, fromNumber string) *TwilioNotifier {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &TwilioNotifier{
		FromNumber: fromNumber,
		client:     client,
	}
}

func (t *TwilioNotifier) Send(to, body string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(t.FromNumber)
	params.SetBody(body)

	message, err := t.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("twilio send failed: %w", err)
	}

	// 1. Convert the message struct to a beautiful, readable JSON string
	prettyJSON, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		// Fallback in case JSON marshaling fails
		fmt.Printf("Twilio message sent: SID=%s, Status=%s\n", *message.Sid, *message.Status)
	} else {
		fmt.Printf("Twilio Response Payload:\n%s\n", string(prettyJSON))
	}

	return nil
}
