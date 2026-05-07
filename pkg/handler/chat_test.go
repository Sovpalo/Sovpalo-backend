package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gorilla/websocket"
)

type healthStub struct{}

func (healthStub) Status(ctx context.Context) (string, error) {
	return "ok", nil
}

func (healthStub) SMTPStatus(ctx context.Context) (string, error) {
	return "ok", nil
}

type authStub struct {
	parseToken func(token string) (int, error)
}

func (s authStub) CreateUser(user model.User) (int, error) { return 0, nil }
func (s authStub) ParseToken(token string) (int, error) {
	if s.parseToken != nil {
		return s.parseToken(token)
	}
	return 0, nil
}
func (s authStub) UserExists(email string) (bool, error)        { return false, nil }
func (s authStub) UsernameExists(username string) (bool, error) { return false, nil }
func (s authStub) GetProfile(userID int64) (model.UserProfile, error) {
	return model.UserProfile{}, nil
}
func (s authStub) UpdateAvatar(userID int64, fileName string, fileData []byte) (model.UserProfile, error) {
	return model.UserProfile{}, nil
}
func (s authStub) DeleteAvatar(userID int64) (model.UserProfile, error) {
	return model.UserProfile{}, nil
}
func (s authStub) DeleteUser(userID int64) error                        { return nil }
func (s authStub) SendCodeToEmail(to string, code string) error         { return nil }
func (s authStub) GenerateCode() string                                 { return "" }
func (s authStub) GenerateToken(email, password string) (string, error) { return "", nil }
func (s authStub) SignIn(input model.SignInInput) (string, error)       { return "", nil }
func (s authStub) SignInTelegram(input model.TelegramSignInInput) (string, error) {
	return "", nil
}
func (s authStub) StartRegistration(input model.SignUpInput) error { return nil }
func (s authStub) VerifyRegistration(input model.SignUpVerifyInput) (string, error) {
	return "", nil
}
func (s authStub) ResendRegistrationCode(email string) error                      { return nil }
func (s authStub) StartPasswordReset(email string) error                          { return nil }
func (s authStub) VerifyPasswordReset(input model.ResetPasswordVerifyInput) error { return nil }
func (s authStub) ResendPasswordResetCode(email string) error                     { return nil }
func (s authStub) PendingRegistrationTTL() time.Duration                          { return 0 }

type chatStub struct {
	listFn   func(companyID int64, userID int64, beforeMessageID int64, limit int) ([]model.ChatMessageView, error)
	createFn func(companyID int64, userID int64, input model.ChatMessageCreateInput) (model.ChatMessageView, error)
	unreadFn func(companyID int64, userID int64) (int64, error)
}

func (s chatStub) ListCompanyMessages(companyID int64, userID int64, beforeMessageID int64, limit int) ([]model.ChatMessageView, error) {
	if s.listFn != nil {
		return s.listFn(companyID, userID, beforeMessageID, limit)
	}
	return nil, nil
}

func (s chatStub) CreateCompanyTextMessage(companyID int64, userID int64, input model.ChatMessageCreateInput) (model.ChatMessageView, error) {
	if s.createFn != nil {
		return s.createFn(companyID, userID, input)
	}
	return model.ChatMessageView{}, nil
}

func (s chatStub) CreateCompanyMediaMessage(companyID int64, userID int64, files []service.ChatUploadFile) (model.ChatMessageView, error) {
	return model.ChatMessageView{}, nil
}

func (s chatStub) DeleteCompanyMessage(companyID int64, messageID int64, userID int64) error {
	return nil
}

func (s chatStub) MarkCompanyMessagesRead(companyID int64, userID int64, input model.ChatMarkReadInput) (model.ChatReadResult, error) {
	return model.ChatReadResult{}, nil
}

func (s chatStub) GetCompanyUnreadCount(companyID int64, userID int64) (int64, error) {
	if s.unreadFn != nil {
		return s.unreadFn(companyID, userID)
	}
	return 0, nil
}

func TestListCompanyChatMessagesReturnsForbiddenForNonMember(t *testing.T) {
	svc := &service.Service{
		Authorization: authStub{
			parseToken: func(token string) (int, error) { return 42, nil },
		},
		Chat: chatStub{
			listFn: func(companyID int64, userID int64, beforeMessageID int64, limit int) ([]model.ChatMessageView, error) {
				return nil, errors.New("user is not a member of the company")
			},
		},
	}

	h := NewHandler(healthStub{}, svc)
	router := h.InitRoutes()

	req := httptest.NewRequest(http.MethodGet, "/companies/9001/chat/messages", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusForbidden, w.Code, w.Body.String())
	}
}

func TestCompanyChatWebSocketMessageCreatedIsPersonalizedForRecipient(t *testing.T) {
	readAt := time.Now().UTC()
	svc := &service.Service{
		Authorization: authStub{
			parseToken: func(token string) (int, error) {
				switch token {
				case "sender-token":
					return 1, nil
				case "recipient-token":
					return 2, nil
				default:
					return 0, errors.New("bad token")
				}
			},
		},
		Chat: chatStub{
			createFn: func(companyID int64, userID int64, input model.ChatMessageCreateInput) (model.ChatMessageView, error) {
				return model.ChatMessageView{
					ID:                  100,
					CompanyID:           companyID,
					SenderID:            userID,
					SenderUsername:      "sender",
					Text:                &input.Text,
					Attachments:         []model.ChatAttachment{},
					CreatedAt:           readAt,
					IsReadByCurrentUser: true,
					ReadAt:              &readAt,
				}, nil
			},
			unreadFn: func(companyID int64, userID int64) (int64, error) {
				return 0, nil
			},
		},
	}

	h := NewHandler(healthStub{}, svc)
	server := httptest.NewServer(h.InitRoutes())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/companies/9001/chat/ws?token=recipient-token"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close()

	postBody := []byte(`{"text":"hello realtime"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/companies/9001/chat/messages", bytes.NewReader(postBody))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sender-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}

	var event model.ChatRealtimeEvent
	if err := conn.ReadJSON(&event); err != nil {
		t.Fatalf("read websocket event failed: %v", err)
	}

	if event.Type != "message_created" {
		t.Fatalf("expected message_created, got %s", event.Type)
	}
	if event.Message == nil {
		t.Fatal("expected message payload")
	}
	if event.Message.IsReadByCurrentUser {
		t.Fatal("expected recipient to receive unread message payload")
	}
	if event.Message.ReadAt != nil {
		t.Fatal("expected recipient read_at to be nil")
	}
}
