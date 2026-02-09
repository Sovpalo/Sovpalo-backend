package handler

import (
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}
	/*err := validatePassword(input.PasswordHash)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "incorrect password")
		return
	}*/

	exists, err := h.services.UserExists(input.Email)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "user verification failed")
		return
	}

	if exists {
		newErrorResponse(c, http.StatusBadRequest, "user already exists")
		return
	}
	/*
	storedCode, err := checkCode(c)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	code := h.services.Authorization.GenerateCode()

	if err := h.services.SendCodeToEmail(input.Email, code); err != nil {
		log.Println("failed to send code:", err)
		newErrorResponse(c, http.StatusBadRequest, "failed to send code")
		return
	}

	mu.Lock()
	verificationCodes[input.Email] = VerificationCode{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		UserData:  input,
		Email:     input.Email,
	}
	mu.Unlock()
	*/
	_, err = h.services.Authorization.CreateUser(model.User{
		Email:    input.Email,
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := h.services.Authorization.GenerateToken(input.Email, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}
