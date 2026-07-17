package config

type GoogleConfig struct {
	ApplicationCredentials string `mapstructure:"application_credentials"`
	SheetID                string `mapstructure:"sheet_id"`
	SheetRange             string `mapstructure:"sheet_range"`
}

type TwilioConfig struct {
	AccountSID   string `mapstructure:"account_sid"`
	AuthToken    string `mapstructure:"auth_token"`
	WhatsAppFrom string `mapstructure:"whatsapp_from"`
}

type Config struct {
	Google GoogleConfig `mapstructure:"google"`
	Twilio TwilioConfig `mapstructure:"twilio"`
}
