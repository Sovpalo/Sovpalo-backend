package model

import "time"

type AuthProvider string

const (
	AuthProviderPassword AuthProvider = "password"
	AuthProviderTelegram AuthProvider = "telegram"
	AuthProviderApple    AuthProvider = "apple"
)

type SignUpInput struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpVerifyInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=4,numeric"`
}

type SignInInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TelegramSignInInput struct {
	InitData  string  `json:"init_data,omitempty"`
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Username  *string `json:"username,omitempty"`
	PhotoURL  *string `json:"photo_url,omitempty"`
	AuthDate  int64   `json:"auth_date"`
	Hash      string  `json:"hash"`
}

type AppleSignInInput struct {
	IdentityToken string  `json:"identity_token" binding:"required"`
	Nonce         *string `json:"nonce,omitempty"`
	Email         *string `json:"email,omitempty"`
	GivenName     *string `json:"given_name,omitempty"`
	FamilyName    *string `json:"family_name,omitempty"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordVerifyInput struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=4,numeric"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UserProfile struct {
	Email     *string        `json:"email,omitempty"`
	Username  string         `json:"username"`
	AvatarURL *string        `json:"avatar_url,omitempty"`
	Providers []AuthProvider `json:"providers"`
}

type AuthChallengeType string

const (
	AuthChallengeTypeSignUp        AuthChallengeType = "sign_up"
	AuthChallengeTypePasswordReset AuthChallengeType = "password_reset"
)

type PendingAuthChallenge struct {
	Type         AuthChallengeType `json:"type"`
	Email        string            `json:"email"`
	Username     string            `json:"username,omitempty"`
	PasswordHash string            `json:"password_hash,omitempty"`
	Code         string            `json:"code"`
	ExpiresAt    time.Time         `json:"expires_at"`
}
