package main

import (
	"testing"
)

func TestGoogleSheetsRepository_parseRows(t *testing.T) {
	repo := &GoogleSheetsRepository{}

	t.Run("well-formed rows parse correctly", func(t *testing.T) {
		rows := [][]interface{}{
			{"  Name  ", " Phone ", " Date ", " Time "},
			{"John Doe", "+254712345678", "2026-07-20", "18:00"},
			{"Jane Smith", "+254798765432", "2026-07-27", "18:00"},
		}

		sessions := repo.parseRows(rows)

		if len(sessions) != 2 {
			t.Fatalf("expected 2 sessions, got %d", len(sessions))
		}
		if sessions[0].Name != "John Doe" || sessions[0].Phone != "+254712345678" {
			t.Errorf("unexpected first session: %+v", sessions[0])
		}
		if sessions[0].Time != "18:00" {
			t.Errorf("expected time to be parsed, got %q", sessions[0].Time)
		}
	})

	t.Run("row with too few columns is skipped, not fatal", func(t *testing.T) {
		rows := [][]interface{}{
			{"John Doe", "+254712345678", "2026-07-20"}, // missing time
			{"Jane Smith", "+254798765432", "2026-07-27", "18:00"},
		}

		sessions := repo.parseRows(rows)

		if len(sessions) != 1 {
			t.Fatalf("expected malformed row to be skipped, got %d sessions", len(sessions))
		}
		if sessions[0].Name != "Jane Smith" {
			t.Errorf("expected remaining valid row to still parse, got: %+v", sessions[0])
		}
	})

	t.Run("row with invalid date is skipped, not fatal", func(t *testing.T) {
		rows := [][]interface{}{
			{"John Doe", "+254712345678", "not-a-date", "18:00"},
			{"Jane Smith", "+254798765432", "2026-07-27", "18:00"},
		}

		sessions := repo.parseRows(rows)

		if len(sessions) != 1 {
			t.Fatalf("expected row with bad date to be skipped, got %d sessions", len(sessions))
		}
	})

	t.Run("empty input returns empty slice without panic", func(t *testing.T) {
		sessions := repo.parseRows([][]interface{}{})

		if len(sessions) != 0 {
			t.Errorf("expected empty result, got %d sessions", len(sessions))
		}
	})

	t.Run("whitespace in fields is trimmed", func(t *testing.T) {
		rows := [][]interface{}{
			{"  Name  ", " Phone ", " Date ", " Time "},
			{"  John Doe  ", " +254712345678 ", " 2026-07-20 ", " 18:00 "},
		}

		sessions := repo.parseRows(rows)

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}
		if sessions[0].Name != "John Doe" {
			t.Errorf("expected trimmed name, got %q", sessions[0].Name)
		}
	})
}
