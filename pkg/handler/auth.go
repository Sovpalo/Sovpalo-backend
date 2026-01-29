package email

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Sovpalo/sovpalo-backend/pkg/handler"
	"github.com/Sovpalo/sovpalo-backend/pkg/model"
)

type AuthEmail struct {
	smtpHost    string
	smtpPort    string
	senderEmail string
	senderPass  string
}

type VerificationCode struct {
	Code       string
	ExpiresAt  time.Time
	UserData   model.User
	Email      string
	IsVerified bool
}

var (
	verificationCodes = make(map[string]VerificationCode)
	mu                sync.Mutex
)

type signInInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *handler.Handler) signUp(c *gin.Context) {
	var input model.User
	if err := c.BindJSON(&input); err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}
	err := validatePassword(input.PasswordHash)
	if err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "incorrect password")
		return
	}

	exists, err := h.services.UserExists(input.Email)
	if err != nil {
		h.newErrorResponse(c, http.StatusBadRequest, "user verification failed")
		return
	}

	if exists {
		h.newErrorResponse(c, http.StatusBadRequest, "user already exists")
		return
	}

	code := h.services.Authorization.GenerateCode()

	if err := h.services.SendCodeToEmail(input.Email, code); err != nil {
		log.Println("failed to send code:", err)
		h.newErrorResponse(c, http.StatusBadRequest, "failed to send code")
		return
	}
}
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password is too short")
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
		return errors.New("invalid password")
	}

	return nil
}
