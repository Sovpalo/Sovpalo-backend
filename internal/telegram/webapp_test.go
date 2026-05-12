package telegram

import (
	"strings"
	"testing"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
)

func TestRenderWebAppHTMLUsesDeepLinkSettings(t *testing.T) {
	html := RenderWebAppHTML("sovpalo", "telegram-auth")

	if !strings.Contains(html, `scheme + "://" + host + "?init_data="`) {
		t.Fatalf("expected deep link builder in html")
	}
	if strings.Contains(html, "__DEEP_LINK_SCHEME__") {
		t.Fatal("expected scheme placeholder to be replaced")
	}
	if strings.Contains(html, "__DEEP_LINK_HOST__") {
		t.Fatal("expected host placeholder to be replaced")
	}
}

func TestRegisterLinksFromConfig(t *testing.T) {
	links := RegisterLinksFromConfig(config.Config{
		TelegramPublicBaseURL:     "http://example.com",
		TelegramBotUsername:       "SovpaloBestBot",
		TelegramMiniAppStartParam: "auth",
		TelegramDeepLinkScheme:    "sovpalo",
		TelegramDeepLinkHost:      "telegram-auth",
	})

	if links.BotURL != "https://t.me/SovpaloBestBot" {
		t.Fatalf("unexpected bot url: %s", links.BotURL)
	}
	if links.MiniAppURL != "https://t.me/SovpaloBestBot?startapp=auth" {
		t.Fatalf("unexpected mini app url: %s", links.MiniAppURL)
	}
	if links.WebAppURL != "http://example.com/telegram/webapp" {
		t.Fatalf("unexpected webapp url: %s", links.WebAppURL)
	}
	if links.DeepLink != "sovpalo://telegram-auth" {
		t.Fatalf("unexpected deep link: %s", links.DeepLink)
	}
}

func TestIsStartCommand(t *testing.T) {
	if !isStartCommand("/start") {
		t.Fatal("expected /start to match")
	}
	if !isStartCommand("/start auth") {
		t.Fatal("expected /start auth to match")
	}
	if isStartCommand("/help") {
		t.Fatal("expected /help not to match")
	}
}
