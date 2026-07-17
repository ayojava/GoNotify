package main

import (
	"testing"
	"time"
)

func TestSession_DaysUntil(t *testing.T) {
	tests := []struct {
		name     string
		session  Session
		from     time.Time
		expected int
	}{
		{
			name:     "same calendar day returns zero",
			session:  Session{Date: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)},
			from:     time.Date(2026, 7, 20, 9, 30, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "three days ahead returns three",
			session:  Session{Date: time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)},
			from:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
			expected: 3,
		},
		{
			name:     "session in the past returns negative",
			session:  Session{Date: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)},
			from:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
			expected: -5,
		},
		{
			name:     "time-of-day is ignored, only calendar day matters",
			session:  Session{Date: time.Date(2026, 7, 21, 23, 59, 0, 0, time.UTC)},
			from:     time.Date(2026, 7, 21, 0, 1, 0, 0, time.UTC),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.DaysUntil(tt.from)
			if got != tt.expected {
				t.Errorf("DaysUntil() = %d, want %d", got, tt.expected)
			}
		})
	}
}
