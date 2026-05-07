package service

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/golang-jwt/jwt/v5"
)

type APNSPushSender struct {
	keyID       string
	teamID      string
	bundleID    string
	privateKey  *ecdsa.PrivateKey
	host        string
	httpClient  *http.Client
	tokenRepo   repository.Notification
	mu          sync.Mutex
	bearerToken string
	tokenExpiry time.Time
}

func NewAPNSPushSender(cfg config.Config, tokenRepo repository.Notification) *APNSPushSender {
	privateKeyPEM := strings.TrimSpace(cfg.APNSPrivateKey)
	if privateKeyPEM == "" && cfg.APNSPrivateKeyPath != "" {
		data, err := os.ReadFile(cfg.APNSPrivateKeyPath)
		if err == nil {
			privateKeyPEM = string(data)
		}
	}

	var privateKey *ecdsa.PrivateKey
	if privateKeyPEM != "" {
		key, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKeyPEM))
		if err == nil {
			privateKey = key
		}
	}

	host := "https://api.sandbox.push.apple.com"
	if cfg.APNSProduction {
		host = "https://api.push.apple.com"
	}

	return &APNSPushSender{
		keyID:      cfg.APNSKeyID,
		teamID:     cfg.APNSTeamID,
		bundleID:   cfg.APNSBundleID,
		privateKey: privateKey,
		host:       host,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		tokenRepo:  tokenRepo,
	}
}

func (s *APNSPushSender) Enabled() bool {
	return s != nil && s.keyID != "" && s.teamID != "" && s.bundleID != "" && s.privateKey != nil
}

func (s *APNSPushSender) Send(token model.PushDeviceToken, notification model.PushNotification) error {
	if !s.Enabled() {
		return nil
	}

	bearer, err := s.authorizationToken()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"aps": map[string]interface{}{
			"alert": map[string]string{
				"title": notification.Title,
				"body":  notification.Message,
			},
			"sound": "default",
		},
		"type": notification.Type,
	}
	for key, value := range notification.Data {
		payload[key] = value
	}
	if notification.RelatedEntityType != nil {
		payload["related_entity_type"] = *notification.RelatedEntityType
	}
	if notification.RelatedEntityID != nil {
		payload["related_entity_id"] = fmt.Sprintf("%d", *notification.RelatedEntityID)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	deviceToken := strings.TrimSpace(token.Token)
	url := fmt.Sprintf("%s/3/device/%s", s.host, deviceToken)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "bearer "+bearer)
	req.Header.Set("apns-topic", s.bundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("content-type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode == http.StatusGone && s.tokenRepo != nil {
		_ = s.tokenRepo.RemovePushDeviceToken(deviceToken)
	}
	return fmt.Errorf("apns returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}

func (s *APNSPushSender) authorizationToken() (string, error) {
	if !s.Enabled() {
		return "", errors.New("apns is not configured")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.bearerToken != "" && now.Before(s.tokenExpiry) {
		return s.bearerToken, nil
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": s.teamID,
		"iat": now.Unix(),
	})
	token.Header["kid"] = s.keyID

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}

	s.bearerToken = signed
	s.tokenExpiry = now.Add(45 * time.Minute)
	return signed, nil
}
