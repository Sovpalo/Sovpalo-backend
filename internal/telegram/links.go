package telegram

import (
	"fmt"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
)

type RegisterLinks struct {
	BotURL     string `json:"bot_url"`
	MiniAppURL string `json:"mini_app_url"`
	WebAppURL  string `json:"webapp_url"`
	DeepLink   string `json:"deep_link"`
}

func RegisterLinksFromConfig(cfg config.Config) RegisterLinks {
	username := strings.TrimPrefix(strings.TrimSpace(cfg.TelegramBotUsername), "@")
	webAppURL := cfg.TelegramWebAppURL()
	links := RegisterLinks{
		WebAppURL: webAppURL,
		DeepLink:  fmt.Sprintf("%s://%s", cfg.TelegramDeepLinkSchemeValue(), cfg.TelegramDeepLinkHostValue()),
	}
	if username != "" {
		links.BotURL = fmt.Sprintf("https://t.me/%s", username)
		startParam := strings.TrimSpace(cfg.TelegramMiniAppStartParam)
		if startParam != "" {
			links.MiniAppURL = fmt.Sprintf("https://t.me/%s?startapp=%s", username, startParam)
		}
	}
	return links
}
