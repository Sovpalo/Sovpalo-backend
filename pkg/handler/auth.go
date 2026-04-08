package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "user_id"
)

func (h *Handler) userIdentity(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)

	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "No authorization header")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		newErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header")
		return
	}
	userId, err := h.services.Authorization.ParseToken(headerParts[1])
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.Set(userCtx, userId)
	c.Next()
}

func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return 0, errors.New("user id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		return 0, errors.New("user id is of invalid type")
	}
	return idInt, nil
}

func mapRegistrationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPassword):
		newErrorResponse(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrUserAlreadyExists), errors.Is(err, service.ErrUsernameAlreadyExists):
		newErrorResponse(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		newErrorResponse(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrPendingRegistrationNotFound), errors.Is(err, service.ErrVerificationCodeExpired):
		newErrorResponse(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrIncorrectVerificationCode):
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
	default:
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
}
