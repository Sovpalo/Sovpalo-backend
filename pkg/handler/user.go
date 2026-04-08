package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

const maxAvatarUploadSize = 5 << 20

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

func (h *Handler) deleteCurrentUser(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.services.Authorization.DeleteUser(int64(userID)); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user deleted",
	})
}

func (h *Handler) uploadCurrentUserAvatar(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	if fileHeader.Size > maxAvatarUploadSize {
		newErrorResponse(c, http.StatusBadRequest, service.ErrAvatarTooLarge.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "failed to open avatar file")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(io.LimitReader(file, maxAvatarUploadSize+1))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "failed to read avatar file")
		return
	}
	if len(fileData) > maxAvatarUploadSize {
		newErrorResponse(c, http.StatusBadRequest, service.ErrAvatarTooLarge.Error())
		return
	}

	profile, err := h.services.Authorization.UpdateAvatar(int64(userID), fileHeader.Filename, fileData)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			newErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrAvatarTooLarge), errors.Is(err, service.ErrAvatarInvalidType):
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) deleteCurrentUserAvatar(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	profile, err := h.services.Authorization.DeleteAvatar(int64(userID))
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
