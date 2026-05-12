package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const telegramAPIBase = "https://api.telegram.org/bot"

type APIClient struct {
	token      string
	httpClient *http.Client
}

func NewAPIClient(token string) *APIClient {
	return &APIClient{
		token: strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: 70 * time.Second,
		},
	}
}

func (c *APIClient) Enabled() bool {
	return c != nil && c.token != ""
}

type BotUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Update struct {
	UpdateID int64   `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Chat      Chat   `json:"chat"`
}

type Chat struct {
	ID int64 `json:"id"`
}

func (c *APIClient) GetMe(ctx context.Context) (BotUser, error) {
	var response struct {
		OK     bool    `json:"ok"`
		Result BotUser `json:"result"`
	}
	if err := c.call(ctx, "getMe", nil, &response); err != nil {
		return BotUser{}, err
	}
	if !response.OK {
		return BotUser{}, fmt.Errorf("telegram getMe failed")
	}
	return response.Result, nil
}

func (c *APIClient) DeleteWebhook(ctx context.Context) error {
	var response struct {
		OK bool `json:"ok"`
	}
	return c.call(ctx, "deleteWebhook", map[string]any{"drop_pending_updates": false}, &response)
}

func (c *APIClient) GetUpdates(ctx context.Context, offset int64, timeout int) ([]Update, error) {
	var response struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	params := map[string]any{
		"offset":  offset,
		"timeout": timeout,
		"allowed_updates": []string{
			"message",
		},
	}
	if err := c.call(ctx, "getUpdates", params, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("telegram getUpdates failed")
	}
	return response.Result, nil
}

func (c *APIClient) SendStartMessage(ctx context.Context, chatID int64, text string, webAppURL string) error {
	replyMarkup := map[string]any{
		"inline_keyboard": [][]map[string]any{
			{
				{
					"text":    "Войти в Sovpalo",
					"web_app": map[string]string{"url": webAppURL},
				},
			},
		},
	}
	params := map[string]any{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": replyMarkup,
	}
	var response struct {
		OK bool `json:"ok"`
	}
	if err := c.call(ctx, "sendMessage", params, &response); err != nil {
		return err
	}
	if !response.OK {
		return fmt.Errorf("telegram sendMessage failed")
	}
	return nil
}

func (c *APIClient) call(ctx context.Context, method string, params map[string]any, dest any) error {
	var body io.Reader
	if params != nil {
		encoded, err := json.Marshal(params)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, telegramAPIBase+c.token+"/"+method, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram %s returned %d: %s", method, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if err := json.Unmarshal(respBody, dest); err != nil {
		return err
	}
	return nil
}
