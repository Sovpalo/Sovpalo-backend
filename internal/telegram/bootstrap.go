package telegram

import (
	"context"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
)

func ResolveBotUsername(ctx context.Context, cfg config.Config) string {
	username := strings.TrimPrefix(strings.TrimSpace(cfg.TelegramBotUsername), "@")
	if username != "" || strings.TrimSpace(cfg.TelegramBotToken) == "" {
		return username
	}

	client := NewAPIClient(cfg.TelegramBotToken)
	me, err := client.GetMe(ctx)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(me.Username)
}
