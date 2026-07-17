package main

import (
	"GoNotify/internal/config"
	"fmt"
	"os"
	"time"
)

// DefaultConfigPath is used by main.go in normal operation.
const DefaultConfigPath = "secrets/application.yaml"

func main() {
	cfg, err := config.Load(DefaultConfigPath)
	if err != nil {
		fmt.Println("Config error:", err)
		os.Exit(1)
	}

	repo := NewGoogleSheetsRepository(cfg.Google.ApplicationCredentials, cfg.Google.SheetID, cfg.Google.SheetRange)
	builder := NewReminderMessageBuilder()
	notifier := NewTwilioNotifier(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken, cfg.Twilio.WhatsAppFrom)

	service := NewReminderService(repo, builder, notifier, []int{1, 0})

	sent, err := service.Run(time.Now())
	if err != nil {
		fmt.Println("Error running reminder service:", err)
		os.Exit(1)
	}

	fmt.Printf("Done. %d reminder(s) sent.\n", sent)
}
