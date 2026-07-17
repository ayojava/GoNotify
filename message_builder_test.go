package main

import (
	"strings"
	"testing"
	"time"
)

func TestReminderMessageBuilder_Build(t *testing.T) {
	builder := NewReminderMessageBuilder()
	session := Session{
		Name: "John Doe",
		Date: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC), // a Monday
	}

	t.Run("same-day message mentions TODAY", func(t *testing.T) {
		msg := builder.Build(session, 0)

		if !strings.Contains(msg, "TODAY") {
			t.Errorf("expected same-day message to contain TODAY, got: %s", msg)
		}
		if !strings.Contains(msg, "John Doe") {
			t.Errorf("expected message to contain the leader's name, got: %s", msg)
		}
	})

	t.Run("upcoming message states days remaining", func(t *testing.T) {
		msg := builder.Build(session, 3)

		if !strings.Contains(msg, "in 3 day(s)") {
			t.Errorf("expected message to state days remaining, got: %s", msg)
		}
		if strings.Contains(msg, "TODAY") {
			t.Errorf("did not expect TODAY wording for a future session, got: %s", msg)
		}
	})

	t.Run("message includes formatted date", func(t *testing.T) {
		msg := builder.Build(session, 3)

		if !strings.Contains(msg, "Monday, Jul 20") {
			t.Errorf("expected message to contain formatted date, got: %s", msg)
		}
	})
}
