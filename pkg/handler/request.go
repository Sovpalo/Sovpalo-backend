package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func publicRequestBaseURL(c *gin.Context) string {
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return ""
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(forwarded, ",")[0]))
	}

	return scheme + "://" + host
}
