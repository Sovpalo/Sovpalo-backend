package handler

import (
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signIn(c *gin.Context) {
	var input model.SignInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	token, err := h.services.Authorization.SignIn(input)
	if err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *Handler) forgotPassword(c *gin.Context) {
	var input model.ForgotPasswordInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Authorization.StartPasswordReset(input.Email); err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":        "verification code sent",
		"expires_in_sec": int(h.services.Authorization.PendingRegistrationTTL().Seconds()),
	})
}

func (h *Handler) verifyForgotPassword(c *gin.Context) {
	var input model.ResetPasswordVerifyInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Authorization.VerifyPasswordReset(input); err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password updated",
	})
}

func (h *Handler) resendForgotPasswordCode(c *gin.Context) {
	var input model.ForgotPasswordInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, bindingErrorMessage(err))
		return
	}

	if err := h.services.Authorization.ResendPasswordResetCode(input.Email); err != nil {
		mapRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "verification code resent",
		"expires_in_sec": int(h.services.Authorization.PendingRegistrationTTL().Seconds()),
	})
}
