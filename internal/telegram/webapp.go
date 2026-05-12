package telegram

import (
	_ "embed"
	"strings"
)

//go:embed webapp.html
var webAppHTMLTemplate string

func RenderWebAppHTML(deepLinkScheme, deepLinkHost string) string {
	scheme := strings.TrimSpace(deepLinkScheme)
	if scheme == "" {
		scheme = "sovpalo"
	}
	host := strings.TrimSpace(deepLinkHost)
	if host == "" {
		host = "telegram-auth"
	}

	html := strings.ReplaceAll(webAppHTMLTemplate, "__DEEP_LINK_SCHEME__", scheme)
	return strings.ReplaceAll(html, "__DEEP_LINK_HOST__", host)
}
