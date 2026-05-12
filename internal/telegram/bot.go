package telegram

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
)

type Bot struct {
	client    *APIClient
	webAppURL string
}

func NewBot(cfg config.Config) *Bot {
	return &Bot{
		client:    NewAPIClient(cfg.TelegramBotToken),
		webAppURL: cfg.TelegramWebAppURL(),
	}
}

func (b *Bot) Enabled() bool {
	return b != nil && b.client.Enabled() && b.webAppURL != ""
}

func (b *Bot) Run(ctx context.Context) error {
	if !b.Enabled() {
		log.Printf("telegram bot disabled: token or web app url is not configured")
		return nil
	}
	if strings.HasPrefix(strings.ToLower(b.webAppURL), "http://") {
		log.Printf("telegram web app url uses http: %s; Telegram WebApp buttons require https in production", b.webAppURL)
	}

	if err := b.client.DeleteWebhook(ctx); err != nil {
		return err
	}

	me, err := b.client.GetMe(ctx)
	if err != nil {
		return err
	}
	log.Printf("telegram bot polling started for @%s", me.Username)

	var offset int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.client.GetUpdates(ctx, offset, 50)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("telegram getUpdates error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if update.Message == nil {
				continue
			}
			if !isStartCommand(update.Message.Text) {
				continue
			}
			if err := b.client.SendStartMessage(ctx, update.Message.Chat.ID, startMessageText(), b.webAppURL); err != nil {
				log.Printf("telegram sendMessage error: %v", err)
			}
		}
	}
}

func isStartCommand(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	if text == "/start" {
		return true
	}
	return strings.HasPrefix(text, "/start ")
}

func startMessageText() string {
	return "Нажмите «Войти в Sovpalo», чтобы подтвердить вход через Telegram и вернуться в приложение."
}
