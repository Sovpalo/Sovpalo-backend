package config

import "testing"

func TestTelegramWebAppURL(t *testing.T) {
	cfg := Config{
		TelegramPublicBaseURL: "http://example.com",
	}
	if got := cfg.TelegramWebAppURL(); got != "http://example.com/telegram/webapp" {
		t.Fatalf("unexpected webapp url: %s", got)
	}

	cfg = Config{TelegramWebAppURLOverride: "https://example.com/custom"}
	if got := cfg.TelegramWebAppURL(); got != "https://example.com/custom" {
		t.Fatalf("unexpected override webapp url: %s", got)
	}
}

func TestTelegramBotPollingEnabled(t *testing.T) {
	cfg := Config{
		TelegramBotEnabled:    true,
		TelegramBotToken:      "token",
		TelegramPublicBaseURL: "http://example.com",
	}
	if !cfg.TelegramBotPollingEnabled() {
		t.Fatal("expected polling to be enabled")
	}

	cfg.TelegramBotToken = ""
	if cfg.TelegramBotPollingEnabled() {
		t.Fatal("expected polling to be disabled without token")
	}
}
