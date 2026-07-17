package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

const testConfigPath = "../testdata/application.yaml"

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

func TestLoad_AllEnvVarsSet_NoFile(t *testing.T) {
	resetViper(t)

	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "service-account.json")
	t.Setenv("GOOGLE_SHEET_ID", "sheet123")
	t.Setenv("GOOGLE_SHEET_RANGE", "Sheet1!A2:C")
	t.Setenv("TWILIO_ACCOUNT_SID", "ACxxxx")
	t.Setenv("TWILIO_AUTH_TOKEN", "tokenxxxx")
	t.Setenv("TWILIO_WHATSAPP_FROM", "whatsapp:+14155238886")

	cfg, err := Load(testConfigPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Google.SheetID != "sheet123" {
		t.Errorf("expected sheet ID from env var, got %q", cfg.Google.SheetID)
	}
	if cfg.Twilio.AccountSID != "ACxxxx" {
		t.Errorf("expected account SID from env var, got %q", cfg.Twilio.AccountSID)
	}
}

func TestLoad_MissingField_ReturnsDescriptiveError(t *testing.T) {
	resetViper(t)

	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "service-account.json")
	t.Setenv("GOOGLE_SHEET_RANGE", "Sheet1!A2:C")
	t.Setenv("TWILIO_ACCOUNT_SID", "ACxxxx")
	t.Setenv("TWILIO_AUTH_TOKEN", "tokenxxxx")
	t.Setenv("TWILIO_WHATSAPP_FROM", "whatsapp:+14155238886")
	// Note: GOOGLE_SHEET_ID is intentionally not set
	// Clear it from the secret file by using it (file has default value but env var should override)

	_, err := Load("../testdata/application-missing-sheet-id.yaml")
	if err == nil {
		t.Fatal("expected an error due to missing google.sheet_id, got nil")
	}
	if !strings.Contains(err.Error(), "google.sheet_id") {
		t.Errorf("expected error to name the missing field, got: %v", err)
	}
}

func TestLoad_NoFileNoEnvVars_ReturnsError(t *testing.T) {
	resetViper(t)
	// Don't set any env vars

	_, err := Load("/nonexistent/secret.yaml")
	if err == nil {
		t.Fatal("expected an error when nothing is configured, got nil")
	}
}

func TestLoad_EnvVarOverridesConfigFile(t *testing.T) {
	resetViper(t)

	viper.SetConfigFile("../testdata/application.yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	t.Setenv("GOOGLE_SHEET_ID", "from-env-sheet-id")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read testdata secret: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Google.SheetID != "from-env-sheet-id" {
		t.Errorf("expected env var to override file value, got %q", cfg.Google.SheetID)
	}
}
