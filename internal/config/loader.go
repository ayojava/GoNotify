package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Load reads secret from configPath (if it exists) plus environment
// variables, and validates the result. Pass DefaultConfigPath in
// production; tests can pass a path under testdata/ instead.
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// AutomaticEnv only overrides keys Viper already knows about — normally
	// learned from the config file. When the file is absent (the CI case
	// this function is designed for), Viper has no keys to check env vars
	// against and Unmarshal silently comes back empty, so each key must be
	// bound explicitly to guarantee env vars are picked up either way.
	for _, key := range []string{
		"google.application_credentials",
		"google.sheet_id",
		"google.sheet_range",
		"twilio.account_sid",
		"twilio.auth_token",
		"twilio.whatsapp_from",
	} {
		if err := viper.BindEnv(key); err != nil {
			return nil, fmt.Errorf("binding env var for %s: %w", key, err)
		}
	}

	// The secret file is optional — in CI (GitHub Actions) there is no
	// application.yaml, only env vars from secrets. Only fail if
	// the file exists but is malformed.
	//
	// Note: viper.ConfigFileNotFoundError is only returned when Viper
	// searches for a config by name; since we set an explicit path via
	// SetConfigFile, a missing file instead surfaces as a raw *fs.PathError,
	// so we also need to check os.IsNotExist directly.
	if err := viper.ReadInConfig(); err != nil {
		_, isViperNotFound := err.(viper.ConfigFileNotFoundError)
		if !isViperNotFound && !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading secret file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshalling secret: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validating secret: %w", err)
	}

	return &config, nil
}

// validate ensures required fields ended up populated, whether from the
// file or from env vars, catching misconfiguration early with a clear
// error instead of failing deep inside the Sheets or Twilio clients.
func (c *Config) validate() error {
	var missing []string
	if c.Google.ApplicationCredentials == "" {
		missing = append(missing, "google.application_credentials")
	}
	if c.Google.SheetID == "" {
		missing = append(missing, "google.sheet_id")
	}
	if c.Google.SheetRange == "" {
		missing = append(missing, "google.sheet_range")
	}
	if c.Twilio.AccountSID == "" {
		missing = append(missing, "twilio.account_sid")
	}
	if c.Twilio.AuthToken == "" {
		missing = append(missing, "twilio.auth_token")
	}
	if c.Twilio.WhatsAppFrom == "" {
		missing = append(missing, "twilio.whatsapp_from")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required secret values: %v", missing)
	}
	return nil
}
