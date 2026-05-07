package service

import (
	"errors"
	"log"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

var (
	ErrPushTokenRequired   = errors.New("push token is required")
	ErrPushPlatformInvalid = errors.New("push platform must be ios")
)

type PushSender interface {
	Send(token model.PushDeviceToken, notification model.PushNotification) error
	Enabled() bool
}

type NotificationService struct {
	repo   repository.Notification
	sender PushSender
}

func NewNotificationService(repo repository.Notification, sender PushSender) *NotificationService {
	return &NotificationService{repo: repo, sender: sender}
}

func (s *NotificationService) RegisterPushToken(userID int64, input model.PushTokenRegisterInput) error {
	token := strings.TrimSpace(input.Token)
	if token == "" {
		return ErrPushTokenRequired
	}
	if input.Platform != "ios" {
		return ErrPushPlatformInvalid
	}
	return s.repo.UpsertPushDeviceToken(userID, token, input.Platform)
}

func (s *NotificationService) DeletePushToken(userID int64, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrPushTokenRequired
	}
	return s.repo.DeletePushDeviceToken(userID, token)
}

func (s *NotificationService) Dispatch(notification model.PushNotification) {
	if err := s.repo.CreateNotification(notification); err != nil {
		log.Printf("notification save error: %v", err)
	}

	if s.sender == nil || !s.sender.Enabled() {
		return
	}

	tokens, err := s.repo.ListPushTokens(notification.UserID)
	if err != nil {
		log.Printf("push token list error: %v", err)
		return
	}

	for _, token := range tokens {
		if err := s.sender.Send(token, notification); err != nil {
			log.Printf("push send error: %v", err)
		}
	}
}
