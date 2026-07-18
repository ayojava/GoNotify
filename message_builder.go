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
			"Hi %s! 🙏 A reminder that you're leading TODAY's (%s) Prayer session from (%s). ",
			s.Name, dateStr, s.Time,
		)
	}

	return fmt.Sprintf(
		"Hi %s! 🙏 Reminder that you're leading the prayer session on %s (in %d day(s)).Incase you cant make it, kindly swap with someone and inform prayer department.",
		s.Name, dateStr, daysUntil,
	)
}
