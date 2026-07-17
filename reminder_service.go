package main

import (
	"fmt"
	"time"
)

type ReminderService struct {
	Repository     ScheduleRepository
	MessageBuilder MessageBuilder
	Notifier       Notifier
	ReminderDays   []int
}

func NewReminderService(repo ScheduleRepository, builder MessageBuilder, notifier Notifier, reminderDays []int) *ReminderService {
	return &ReminderService{
		Repository:     repo,
		MessageBuilder: builder,
		Notifier:       notifier,
		ReminderDays:   reminderDays,
	}
}

func (svc *ReminderService) Run(now time.Time) (int, error) {
	sessions, err := svc.Repository.LoadSchedule()
	if err != nil {
		return 0, fmt.Errorf("loading schedule: %w", err)
	}

	sent := 0
	for _, s := range sessions {
		daysUntil := s.DaysUntil(now)

		if !svc.isReminderDay(daysUntil) {
			continue
		}

		msg := svc.MessageBuilder.Build(s, daysUntil)

		if err := svc.Notifier.Send(s.Phone, msg); err != nil {
			fmt.Printf("Failed to message %s: %v\n", s.Name, err)
			continue
		}

		fmt.Printf("Reminder sent to %s (%s) for session on %s\n", s.Name, s.Phone, s.Date.Format("Jan 2, 2006"))
		sent++
	}

	return sent, nil
}

func (svc *ReminderService) isReminderDay(daysUntil int) bool {
	for _, d := range svc.ReminderDays {
		if daysUntil == d {
			return true
		}
	}
	return false
}
