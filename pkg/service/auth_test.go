package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
)

type authRepoStub struct {
	userByTelegramID map[int64]model.User
	takenUsernames   map[string]bool
	createdUsers     []model.User
	nextUserID       int
}

func (s *authRepoStub) UserExists(email string) (bool, error) { return false, nil }

func (s *authRepoStub) UsernameExists(username string) (bool, error) {
	return s.takenUsernames[username], nil
}

func (s *authRepoStub) CreateUser(user model.User) (int, error) {
	s.createdUsers = append(s.createdUsers, user)
	s.nextUserID++
	if s.userByTelegramID == nil {
		s.userByTelegramID = map[int64]model.User{}
	}
	if user.TelegramID != nil {
		user.ID = int64(s.nextUserID)
		s.userByTelegramID[*user.TelegramID] = user
	}
	return s.nextUserID, nil
}

func (s *authRepoStub) GetUser(email, password string) (model.User, error) {
	return model.User{}, pgx.ErrNoRows
}

func (s *authRepoStub) GetUserByEmail(email string) (model.User, error) {
	return model.User{}, pgx.ErrNoRows
}

func (s *authRepoStub) GetUserByTelegramID(telegramID int64) (model.User, error) {
	user, ok := s.userByTelegramID[telegramID]
	if !ok {
		return model.User{}, pgx.ErrNoRows
	}
	return user, nil
}

func (s *authRepoStub) GetUserByID(userID int64) (model.User, error) {
	return model.User{}, pgx.ErrNoRows
}

func (s *authRepoStub) UpdateUserAvatar(userID int64, avatarURL *string) error { return nil }
func (s *authRepoStub) DeleteUser(userID int64) error                          { return nil }
func (s *authRepoStub) UpdateUserPassword(email string, passwordHash string) error {
	return nil
}
func (s *authRepoStub) SavePendingAuthChallenge(challenge model.PendingAuthChallenge, ttl time.Duration) error {
	return nil
}
func (s *authRepoStub) GetPendingAuthChallenge(challengeType model.AuthChallengeType, email string) (model.PendingAuthChallenge, error) {
	return model.PendingAuthChallenge{}, errors.New("not implemented")
}
func (s *authRepoStub) DeletePendingAuthChallenge(challengeType model.AuthChallengeType, email string) error {
	return nil
}

func TestAuthServiceSignInTelegramCreatesUser(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "telegram-bot-token")

	repo := &authRepoStub{
		userByTelegramID: map[int64]model.User{},
		takenUsernames:   map[string]bool{"alice": true},
	}
	svc := NewAuthService(repo)

	input := model.TelegramSignInInput{
		ID:        777,
		FirstName: "Alice",
		Username:  strPtr("alice"),
		AuthDate:  time.Now().Unix(),
	}
	input.Hash = telegramAuthHash(input, os.Getenv("TELEGRAM_BOT_TOKEN"))

	token, err := svc.SignInTelegram(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if len(repo.createdUsers) != 1 {
		t.Fatalf("expected 1 created user, got %d", len(repo.createdUsers))
	}
	if repo.createdUsers[0].Username != "alice_1" {
		t.Fatalf("expected username alice_1, got %s", repo.createdUsers[0].Username)
	}
	if repo.createdUsers[0].Email != nil {
		t.Fatalf("expected nil email, got %#v", repo.createdUsers[0].Email)
	}
	if repo.createdUsers[0].TelegramID == nil || *repo.createdUsers[0].TelegramID != 777 {
		t.Fatalf("unexpected telegram id %#v", repo.createdUsers[0].TelegramID)
	}
}

func TestAuthServiceSignInTelegramAcceptsWebAppInitData(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "telegram-bot-token")

	repo := &authRepoStub{
		userByTelegramID: map[int64]model.User{},
		takenUsernames:   map[string]bool{},
	}
	svc := NewAuthService(repo)

	initData := telegramWebAppInitData(
		`{"id":888,"first_name":"Bob","username":"bob_tg","photo_url":"https://t.me/i/userpic/320/bob.jpg"}`,
		time.Now().Unix(),
		os.Getenv("TELEGRAM_BOT_TOKEN"),
	)

	token, err := svc.SignInTelegram(model.TelegramSignInInput{InitData: initData})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if len(repo.createdUsers) != 1 {
		t.Fatalf("expected 1 created user, got %d", len(repo.createdUsers))
	}
	if repo.createdUsers[0].Username != "bob_tg" {
		t.Fatalf("expected username bob_tg, got %s", repo.createdUsers[0].Username)
	}
	if repo.createdUsers[0].TelegramID == nil || *repo.createdUsers[0].TelegramID != 888 {
		t.Fatalf("unexpected telegram id %#v", repo.createdUsers[0].TelegramID)
	}
}

func TestAuthServiceSignInTelegramRejectsInvalidHash(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "telegram-bot-token")

	svc := NewAuthService(&authRepoStub{})
	input := model.TelegramSignInInput{
		ID:        777,
		FirstName: "Alice",
		AuthDate:  time.Now().Unix(),
		Hash:      "bad-hash",
	}

	_, err := svc.SignInTelegram(input)
	if !errors.Is(err, ErrInvalidTelegramAuth) {
		t.Fatalf("expected %v, got %v", ErrInvalidTelegramAuth, err)
	}
}

func TestAuthServiceSignInTelegramRejectsInvalidWebAppHash(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "telegram-bot-token")

	svc := NewAuthService(&authRepoStub{})
	initData := telegramWebAppInitData(
		`{"id":888,"first_name":"Bob","username":"bob_tg"}`,
		time.Now().Unix(),
		os.Getenv("TELEGRAM_BOT_TOKEN"),
	) + "bad"

	_, err := svc.SignInTelegram(model.TelegramSignInInput{InitData: initData})
	if !errors.Is(err, ErrInvalidTelegramAuth) {
		t.Fatalf("expected %v, got %v", ErrInvalidTelegramAuth, err)
	}
}

func TestTelegramDataCheckStringSortsKeys(t *testing.T) {
	actual := telegramDataCheckString(map[string]string{
		"username":   "alice",
		"id":         "777",
		"auth_date":  "1710000000",
		"first_name": "Alice",
	})
	expected := "auth_date=1710000000\nfirst_name=Alice\nid=777\nusername=alice"
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func strPtr(value string) *string {
	return &value
}

func telegramWebAppInitData(userJSON string, authDate int64, botToken string) string {
	values := url.Values{}
	values.Set("auth_date", fmt.Sprintf("%d", authDate))
	values.Set("query_id", "AAHdF6IQAAAAAN0XohDhrOrc")
	values.Set("user", userJSON)

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	secret := secretMAC.Sum(nil)

	hashMAC := hmac.New(sha256.New, secret)
	hashMAC.Write([]byte(telegramDataCheckStringFromValues(values)))
	values.Set("hash", hex.EncodeToString(hashMAC.Sum(nil)))

	return values.Encode()
}
