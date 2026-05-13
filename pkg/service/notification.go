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
	repo          repository.Notification
	sender        PushSender
	dispatchDebug bool
}

func NewNotificationService(repo repository.Notification, sender PushSender, dispatchDebug bool) *NotificationService {
	return &NotificationService{repo: repo, sender: sender, dispatchDebug: dispatchDebug}
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

	if s.dispatchDebug {
		suffixes := make([]string, 0, len(tokens))
		for _, t := range tokens {
			suffixes = append(suffixes, pushTokenSuffix(t.Token))
		}
		log.Printf("push dispatch: recipient_user_id=%d notification_type=%s related_entity_type=%v related_entity_id=%v push_token_count=%d token_suffixes=[%s]",
			notification.UserID, notification.Type, notification.RelatedEntityType, notification.RelatedEntityID, len(tokens), strings.Join(suffixes, ","))
	}

	for _, token := range tokens {
		if err := s.sender.Send(token, notification); err != nil {
			log.Printf("push send error: %v", err)
		}
	}
}

func pushTokenSuffix(hexToken string) string {
	t := strings.TrimSpace(hexToken)
	if len(t) <= 8 {
		return t
	}
	return t[len(t)-8:]
}
