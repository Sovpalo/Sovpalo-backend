package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
)

func TestTelegramRegisterLinksUsesRequestHost(t *testing.T) {
	h := NewHandler(healthStub{}, &service.Service{}, config.Config{
		TelegramMiniAppStartParam: "auth",
	}, "SovpaloBestBot")

	router := h.InitRoutes()
	req := httptest.NewRequest(http.MethodGet, "/auth/telegram/register", nil)
	req.Host = "2.56.241.112"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload struct {
		BotURL     string `json:"bot_url"`
		MiniAppURL string `json:"mini_app_url"`
		WebAppURL  string `json:"webapp_url"`
		DeepLink   string `json:"deep_link"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.WebAppURL != "http://2.56.241.112/telegram/webapp" {
		t.Fatalf("unexpected webapp url: %s", payload.WebAppURL)
	}
	if payload.MiniAppURL != "https://t.me/SovpaloBestBot?startapp=auth" {
		t.Fatalf("unexpected mini app url: %s", payload.MiniAppURL)
	}
	if payload.DeepLink != "sovpalo://telegram-auth" {
		t.Fatalf("unexpected deep link: %s", payload.DeepLink)
	}
}
