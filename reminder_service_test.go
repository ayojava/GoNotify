package main

import (
	"errors"
	"testing"
	"time"
)

func TestReminderService_Run(t *testing.T) {
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)

	t.Run("sends reminder for session exactly on a reminder day", func(t *testing.T) {
		repo := &fakeScheduleRepository{
			Sessions: []Session{
				{Name: "John", Phone: "+254700000001", Date: now.AddDate(0, 0, 3)}, // 3 days out
			},
		}
		notifier := &fakeNotifier{}
		svc := NewReminderService(repo, &fakeMessageBuilder{}, notifier, []int{3, 0})

		sent, err := svc.Run(now)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sent != 1 {
			t.Errorf("expected 1 reminder sent, got %d", sent)
		}
		if len(notifier.Sent) != 1 || notifier.Sent[0].To != "+254700000001" {
			t.Errorf("expected notifier to have been called for John, got: %+v", notifier.Sent)
		}
	})

	t.Run("does not send for a session outside the reminder window", func(t *testing.T) {
		repo := &fakeScheduleRepository{
			Sessions: []Session{
				{Name: "John", Phone: "+254700000001", Date: now.AddDate(0, 0, 5)}, // 5 days out, not in {3, 0}
			},
		}
		notifier := &fakeNotifier{}
		svc := NewReminderService(repo, &fakeMessageBuilder{}, notifier, []int{3, 0})

		sent, err := svc.Run(now)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sent != 0 {
			t.Errorf("expected 0 reminders sent, got %d", sent)
		}
		if len(notifier.Sent) != 0 {
			t.Errorf("expected notifier not to be called, got: %+v", notifier.Sent)
		}
	})

	t.Run("only matching sessions trigger sends, others are skipped", func(t *testing.T) {
		repo := &fakeScheduleRepository{
			Sessions: []Session{
				{Name: "John", Phone: "+254700000001", Date: now.AddDate(0, 0, 3)},  // matches
				{Name: "Jane", Phone: "+254700000002", Date: now.AddDate(0, 0, 10)}, // no match
				{Name: "Mary", Phone: "+254700000003", Date: now},                   // matches (same day)
			},
		}
		notifier := &fakeNotifier{}
		svc := NewReminderService(repo, &fakeMessageBuilder{}, notifier, []int{3, 0})

		sent, err := svc.Run(now)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sent != 2 {
			t.Errorf("expected 2 reminders sent, got %d", sent)
		}
	})

	t.Run("one failed send does not stop remaining sessions from being processed", func(t *testing.T) {
		repo := &fakeScheduleRepository{
			Sessions: []Session{
				{Name: "John", Phone: "+254700000001", Date: now}, // will fail
				{Name: "Jane", Phone: "+254700000002", Date: now}, // should still succeed
			},
		}
		notifier := &fakeNotifier{
			ErrFor: map[string]error{
				"+254700000001": errors.New("simulated twilio failure"),
			},
		}
		svc := NewReminderService(repo, &fakeMessageBuilder{}, notifier, []int{3, 0})

		sent, err := svc.Run(now)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sent != 1 {
			t.Errorf("expected 1 successful send despite one failure, got %d", sent)
		}
		if len(notifier.Sent) != 1 || notifier.Sent[0].To != "+254700000002" {
			t.Errorf("expected only Jane to have received a message, got: %+v", notifier.Sent)
		}
	})

	t.Run("repository error is returned and no sends are attempted", func(t *testing.T) {
		repo := &fakeScheduleRepository{
			Err: errors.New("sheets API unreachable"),
		}
		notifier := &fakeNotifier{}
		svc := NewReminderService(repo, &fakeMessageBuilder{}, notifier, []int{3, 0})

		sent, err := svc.Run(now)

		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if sent != 0 {
			t.Errorf("expected 0 sent on repository error, got %d", sent)
		}
		if len(notifier.Sent) != 0 {
			t.Errorf("expected notifier never to be called, got: %+v", notifier.Sent)
		}
	})
}
