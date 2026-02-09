package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/dgrijalva/jwt-go"
	/*
		"crypto/tls"
		"gopkg.in/gomail.v2"
	*/)

type AuthService struct {
	repo         repository.Authorization
	jwtSecret    []byte
	passwordSalt string
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{
		repo:         repo,
		jwtSecret:    []byte(os.Getenv("JWT_SECRET")),
		passwordSalt: os.Getenv("PASSWORD_SALT"),
	}
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

func (s *AuthService) ParseToken(accessToken string) (int, error) {
	if len(s.jwtSecret) == 0 {
		return 0, errors.New("JWT_SECRET not set")
	}
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("Invalid token claims")
	}
	return claims.UserId, nil
}

func (s *AuthService) UserExists(email string) (bool, error) {
	return s.repo.UserExists(email)
}

func (s *AuthService) SendCodeToEmail(to string, code string) error {
	/*
		m := gomail.NewMessage()
		m.SetHeader("From", SMTP_USERNAME)
		m.SetHeader("To", to)
		m.SetHeader("Subject", "Sovpalo - Confirmation Code")
		m.SetBody("text/plain", fmt.Sprintf("Your onetime verification code: %s", code))

		d := gomail.NewDialer(SMPT_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		if err := d.DialAndSend(m); err != nil {
			return fmt.Errorf("error sending email: %w", err)
		}
	*/
	return nil
}

func (s *AuthService) GenerateCode() string {

	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func (s *AuthService) GenerateToken(email, password string) (string, error) {
	if len(s.jwtSecret) == 0 {
		return "", errors.New("JWT_SECRET not set")
	}
	passwordHash, err := s.generatePasswordHash(password)
	if err != nil {
		return "", err
	}
	user, err := s.repo.GetUser(email, passwordHash)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(720 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		int(user.ID),
	})
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generatePasswordHash(password string) (string, error) {
	if s.passwordSalt == "" {
		return "", errors.New("PASSWORD_SALT not set")
	}
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(s.passwordSalt))), nil
}

func (s *AuthService) CreateUser(user model.User) (int, error) {
	passwordHash, err := s.generatePasswordHash(user.Password)
	if err != nil {
		return 0, err
	}
	user.Password = passwordHash
	return s.repo.CreateUser(user)
}
