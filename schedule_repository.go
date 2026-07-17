package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type ScheduleRepository interface {
	LoadSchedule() ([]Session, error)
}

type GoogleSheetsRepository struct {
	CredentialsPath string
	SheetID         string
	SheetRange      string
}

func NewGoogleSheetsRepository(credentialsPath, sheetID, sheetRange string) *GoogleSheetsRepository {
	return &GoogleSheetsRepository{
		CredentialsPath: credentialsPath,
		SheetID:         sheetID,
		SheetRange:      sheetRange,
	}
}

func (r *GoogleSheetsRepository) parseRows(rows [][]interface{}) []Session {
	var sessions []Session

	for i, row := range rows {
		if len(row) < 3 {
			fmt.Printf("Skipping malformed row %d: %v\n", i+1, row)
			continue
		}

		if i == 0 {
			i++
			continue
		}

		name := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		phone := strings.TrimSpace(fmt.Sprintf("%v", row[1]))
		dateStr := strings.TrimSpace(fmt.Sprintf("%v", row[2]))

		date, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			fmt.Printf("Skipping row %d, bad date %q: %v\n", i+1, dateStr, err)
			continue
		}

		fmt.Printf("Parsed row %d: Name=%q, Phone=%q, Date=%s\n", i, name, phone, date.Format(time.RFC3339))
		sessions = append(sessions, Session{
			Name:  name,
			Phone: phone,
			Date:  date,
		})
	}

	return sessions
}

func (r *GoogleSheetsRepository) authenticatedClient(ctx context.Context) (*http.Client, error) {
	credBytes, err := os.ReadFile(r.CredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}

	config, err := google.JWTConfigFromJSON(credBytes, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	return config.Client(ctx), nil
}

func (r *GoogleSheetsRepository) LoadSchedule() ([]Session, error) {
	ctx := context.Background()

	client, err := r.authenticatedClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("authenticating with google: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(r.SheetID, r.SheetRange).Do()
	if err != nil {
		return nil, fmt.Errorf("fetching sheet data: %w", err)
	}

	return r.parseRows(resp.Values), nil
}
