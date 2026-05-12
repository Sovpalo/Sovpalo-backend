package telegram

import (
	"testing"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
)

func TestRegisterLinksUsesRequestBaseURLFallback(t *testing.T) {
	links := RegisterLinksFromConfig(config.Config{
		TelegramMiniAppStartParam: "auth",
	}, RegisterLinkOptions{
		PublicBaseURL: "http://2.56.241.112",
		Username:      "SovpaloBestBot",
	})

	if links.WebAppURL != "http://2.56.241.112/telegram/webapp" {
		t.Fatalf("unexpected webapp url: %s", links.WebAppURL)
	}
	if links.MiniAppURL != "https://t.me/SovpaloBestBot?startapp=auth" {
		t.Fatalf("unexpected mini app url: %s", links.MiniAppURL)
	}
}
