package main

import "fmt"

// fakeScheduleRepository is a test double for ScheduleRepository.
// Configure Sessions and/or Err to control its return value.
type fakeScheduleRepository struct {
	Sessions []Session
	Err      error
}

func (f *fakeScheduleRepository) LoadSchedule() ([]Session, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Sessions, nil
}

// fakeMessageBuilder returns a fixed, predictable message so tests can
// assert on it without depending on ReminderMessageBuilder's real wording.
type fakeMessageBuilder struct{}

func (f *fakeMessageBuilder) Build(s Session, daysUntil int) string {
	return fmt.Sprintf("fake message for %s, %d day(s)", s.Name, daysUntil)
}

// fakeNotifier records every Send call so tests can assert who was
// messaged and with what body. ErrFor lets a specific recipient's send
// fail, to test partial-failure handling in ReminderService.Run.
// This is what ReminderService actually depends on (the Notifier
// interface) — it's the same whether the real implementation is the
// raw HTTP client or the twilio-go SDK, so no test changes were needed
// when notifier.go was rewritten to use the SDK.
type fakeNotifier struct {
	Sent   []sentMessage
	ErrFor map[string]error // phone -> error to return for that recipient
}

type sentMessage struct {
	To   string
	Body string
}

func (f *fakeNotifier) Send(to, body string) error {
	if err, ok := f.ErrFor[to]; ok {
		return err
	}
	f.Sent = append(f.Sent, sentMessage{To: to, Body: body})
	return nil
}
