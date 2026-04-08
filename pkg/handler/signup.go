package handler

import (
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) {
	var input model.SignUpInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Authorization.StartRegistration(input); err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":        "verification code sent",
		"expires_in_sec": int(h.services.Authorization.PendingRegistrationTTL().Seconds()),
	})
}

func (h *Handler) verifySignUp(c *gin.Context) {
	var input model.SignUpVerifyInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	token, err := h.services.Authorization.VerifyRegistration(input)
	if err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *Handler) resendSignUpCode(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Authorization.ResendRegistrationCode(input.Email); err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "verification code resent",
		"expires_in_sec": int(h.services.Authorization.PendingRegistrationTTL().Seconds()),
	})
}
