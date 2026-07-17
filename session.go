package main

import "time"

type Session struct {
	Name  string
	Phone string
	Date  time.Time
}

func (s Session) DaysUntil(from time.Time) int {
	fromDay := truncateToDay(from)
	sessionDay := truncateToDay(s.Date)
	return int(sessionDay.Sub(fromDay).Hours() / 24)
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
