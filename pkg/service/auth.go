package service

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidPassword             = errors.New("invalid password")
	ErrInvalidCredentials          = errors.New("invalid email or password")
	ErrUserAlreadyExists           = errors.New("user with this email already exists")
	ErrUsernameAlreadyExists       = errors.New("user with this username already exists")
	ErrUserNotFound                = errors.New("user not found")
	ErrPendingRegistrationNotFound = errors.New("pending registration not found")
	ErrVerificationCodeExpired     = errors.New("verification code expired")
	ErrIncorrectVerificationCode   = errors.New("incorrect verification code")
)

type AuthService struct {
	repo         repository.Authorization
	jwtSecret    []byte
	passwordSalt string
	pendingTTL   time.Duration
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{
		repo:         repo,
		jwtSecret:    []byte(os.Getenv("JWT_SECRET")),
		passwordSalt: os.Getenv("PASSWORD_SALT"),
		pendingTTL:   10 * time.Minute,
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

func (s *AuthService) UsernameExists(username string) (bool, error) {
	return s.repo.UsernameExists(username)
}

func (s *AuthService) GetProfile(userID int64) (model.UserProfile, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserProfile{}, ErrUserNotFound
		}
		return model.UserProfile{}, err
	}

	return model.UserProfile{
		Email:    user.Email,
		Username: user.Username,
	}, nil
}

func (s *AuthService) SendCodeToEmail(to string, code string) error {
	body := fmt.Sprintf("Your verification code is %s. It expires in %d minutes.", code, int(s.pendingTTL.Minutes()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sendSMTPTextEmail(ctx, to, "Sovpalo verification code", body); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

func (s *AuthService) GenerateCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
	}
	return fmt.Sprintf("%04d", n.Int64())
}

func (s *AuthService) GenerateToken(email, password string) (string, error) {
	if len(s.jwtSecret) == 0 {
		return "", errors.New("JWT_SECRET not set")
	}
	passwordHash, err := s.generatePasswordHash(password)
	if err != nil {
		return "", err
	}
	return s.generateTokenForUser(email, passwordHash)
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

func (s *AuthService) StartRegistration(input model.SignUpInput) error {
	if err := validatePassword(input.Password); err != nil {
		return err
	}

	if err := s.ensureEmailAndUsernameAvailable(input.Email, input.Username); err != nil {
		return err
	}

	passwordHash, err := s.generatePasswordHash(input.Password)
	if err != nil {
		return err
	}

	return s.startChallenge(model.PendingAuthChallenge{
		Type:         model.AuthChallengeTypeSignUp,
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: passwordHash,
	})
}

func (s *AuthService) VerifyRegistration(input model.SignUpVerifyInput) (string, error) {
	challenge, err := s.verifyChallenge(model.AuthChallengeTypeSignUp, input.Email, input.Code)
	if err != nil {
		return "", err
	}

	if err := s.ensureEmailAndUsernameAvailable(challenge.Email, challenge.Username); err != nil {
		return "", err
	}

	if _, err := s.repo.CreateUser(model.User{
		Email:    challenge.Email,
		Username: challenge.Username,
		Password: challenge.PasswordHash,
	}); err != nil {
		return "", err
	}

	if err := s.repo.DeletePendingAuthChallenge(model.AuthChallengeTypeSignUp, input.Email); err != nil {
		return "", err
	}

	return s.generateTokenForUser(challenge.Email, challenge.PasswordHash)
}

func (s *AuthService) ResendRegistrationCode(email string) error {
	return s.resendChallenge(model.AuthChallengeTypeSignUp, email)
}

func (s *AuthService) StartSignIn(input model.SignInInput) error {
	passwordHash, err := s.generatePasswordHash(input.Password)
	if err != nil {
		return err
	}

	if _, err := s.repo.GetUser(input.Email, passwordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	return s.startChallenge(model.PendingAuthChallenge{
		Type:         model.AuthChallengeTypeSignIn,
		Email:        input.Email,
		PasswordHash: passwordHash,
	})
}

func (s *AuthService) VerifySignIn(input model.SignUpVerifyInput) (string, error) {
	challenge, err := s.verifyChallenge(model.AuthChallengeTypeSignIn, input.Email, input.Code)
	if err != nil {
		return "", err
	}

	if err := s.repo.DeletePendingAuthChallenge(model.AuthChallengeTypeSignIn, input.Email); err != nil {
		return "", err
	}

	return s.generateTokenForUser(challenge.Email, challenge.PasswordHash)
}

func (s *AuthService) ResendSignInCode(email string) error {
	return s.resendChallenge(model.AuthChallengeTypeSignIn, email)
}

func (s *AuthService) StartPasswordReset(email string) error {
	if _, err := s.repo.GetUserByEmail(email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return s.startChallenge(model.PendingAuthChallenge{
		Type:  model.AuthChallengeTypePasswordReset,
		Email: email,
	})
}

func (s *AuthService) VerifyPasswordReset(input model.ResetPasswordVerifyInput) error {
	if err := validatePassword(input.NewPassword); err != nil {
		return err
	}

	challenge, err := s.verifyChallenge(model.AuthChallengeTypePasswordReset, input.Email, input.Code)
	if err != nil {
		return err
	}

	passwordHash, err := s.generatePasswordHash(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUserPassword(challenge.Email, passwordHash); err != nil {
		return err
	}

	return s.repo.DeletePendingAuthChallenge(model.AuthChallengeTypePasswordReset, input.Email)
}

func (s *AuthService) ResendPasswordResetCode(email string) error {
	return s.resendChallenge(model.AuthChallengeTypePasswordReset, email)
}

func (s *AuthService) PendingRegistrationTTL() time.Duration {
	return s.pendingTTL
}

func (s *AuthService) generateTokenForUser(email, passwordHash string) (string, error) {
	if len(s.jwtSecret) == 0 {
		return "", errors.New("JWT_SECRET not set")
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

func (s *AuthService) ensureEmailAndUsernameAvailable(email, username string) error {
	emailExists, err := s.repo.UserExists(email)
	if err != nil {
		return err
	}
	if emailExists {
		return ErrUserAlreadyExists
	}

	usernameExists, err := s.repo.UsernameExists(username)
	if err != nil {
		return err
	}
	if usernameExists {
		return ErrUsernameAlreadyExists
	}

	return nil
}

func (s *AuthService) startChallenge(challenge model.PendingAuthChallenge) error {
	challenge.Code = s.GenerateCode()
	challenge.ExpiresAt = time.Now().Add(s.pendingTTL)

	if err := s.repo.SavePendingAuthChallenge(challenge, s.pendingTTL); err != nil {
		return err
	}

	if err := s.SendCodeToEmail(challenge.Email, challenge.Code); err != nil {
		_ = s.repo.DeletePendingAuthChallenge(challenge.Type, challenge.Email)
		return err
	}

	return nil
}

func (s *AuthService) verifyChallenge(challengeType model.AuthChallengeType, email, code string) (model.PendingAuthChallenge, error) {
	challenge, err := s.repo.GetPendingAuthChallenge(challengeType, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.PendingAuthChallenge{}, ErrPendingRegistrationNotFound
		}
		return model.PendingAuthChallenge{}, err
	}

	if time.Now().After(challenge.ExpiresAt) {
		_ = s.repo.DeletePendingAuthChallenge(challengeType, email)
		return model.PendingAuthChallenge{}, ErrVerificationCodeExpired
	}

	if challenge.Code != code {
		return model.PendingAuthChallenge{}, ErrIncorrectVerificationCode
	}

	return challenge, nil
}

func (s *AuthService) resendChallenge(challengeType model.AuthChallengeType, email string) error {
	challenge, err := s.repo.GetPendingAuthChallenge(challengeType, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrPendingRegistrationNotFound
		}
		return err
	}

	challenge.Code = s.GenerateCode()
	challenge.ExpiresAt = time.Now().Add(s.pendingTTL)

	if err := s.repo.SavePendingAuthChallenge(challenge, s.pendingTTL); err != nil {
		return err
	}

	return s.SendCodeToEmail(email, challenge.Code)
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", ErrInvalidPassword)
	}

	hasLower := false
	hasUpper := false
	hasDigit := false

	for _, ch := range password {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}

	if !(hasLower && hasUpper && hasDigit) {
		return fmt.Errorf("%w: password must contain lowercase, uppercase letters and digits", ErrInvalidPassword)
	}

	return nil
}
