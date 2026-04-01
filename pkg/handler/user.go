package handler

import (
	"errors"
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) getCurrentUser(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	profile, err := h.services.Authorization.GetProfile(int64(userID))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, profile)
}
