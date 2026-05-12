package handler

import (
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/internal/telegram"
	"github.com/gin-gonic/gin"
)

func (h *Handler) telegramWebApp(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(telegram.RenderWebAppHTML(
		h.appConfig.TelegramDeepLinkSchemeValue(),
		h.appConfig.TelegramDeepLinkHostValue(),
	)))
}

func (h *Handler) telegramRegisterLinks(c *gin.Context) {
	c.JSON(http.StatusOK, telegram.RegisterLinksFromConfig(h.appConfig))
}
