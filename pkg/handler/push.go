package handler

import (
	"errors"
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) registerPushToken(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var input model.PushTokenRegisterInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Notification.RegisterPushToken(int64(userID), input); err != nil {
		if errors.Is(err, service.ErrPushTokenRequired) || errors.Is(err, service.ErrPushPlatformInvalid) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deletePushToken(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var input struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Notification.DeletePushToken(int64(userID), input.Token); err != nil {
		if errors.Is(err, service.ErrPushTokenRequired) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
