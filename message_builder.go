package main

import "fmt"

type MessageBuilder interface {
	Build(session Session, daysUntil int) string
}

type ReminderMessageBuilder struct{}

func NewReminderMessageBuilder() *ReminderMessageBuilder {
	return &ReminderMessageBuilder{}
}

func (b *ReminderMessageBuilder) Build(s Session, daysUntil int) string {
	dateStr := s.Date.Format("Monday, Jan 2")

	if daysUntil == 0 {
		return fmt.Sprintf(
			"Hi %s! 🙏 Just a reminder that you're leading TODAY's session (%s). Wishing you a great session!",
			s.Name, dateStr,
		)
	}

	return fmt.Sprintf(
		"Hi %s! 🙏 Reminder that you're leading the session on %s (in %d day(s)). Please prepare and let us know if you can't make it.",
		s.Name, dateStr, daysUntil,
	)
}
