package config

import (
	"fmt"
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

	// The secret file is optional — in CI (GitHub Actions) there is no
	// application.yaml, only env vars from secrets. Only fail if
	// the file exists but is malformed.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
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
