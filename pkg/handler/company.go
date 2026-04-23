package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createCompany(c *gin.Context) {
	var input struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		AvatarURL   *string `json:"avatar_url"`
	}
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.services.Company.CreateCompany(int64(userID), input.Name, input.Description, input.AvatarURL)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) listCompanies(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companies, err := h.services.Company.ListCompanies(int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, companies)
}

func (h *Handler) getCompany(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	company, err := h.services.Company.GetCompany(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusNotFound, "company not found")
		return
	}

	c.JSON(http.StatusOK, company)
}

func (h *Handler) updateCompany(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	input, avatarFileName, avatarFileData, err := parseCompanyUpdateInput(c)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Company.UpdateCompany(companyID, int64(userID), input, avatarFileName, avatarFileData); err != nil {
		switch {
		case errors.Is(err, service.ErrAvatarTooLarge), errors.Is(err, service.ErrAvatarInvalidType):
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func parseCompanyUpdateInput(c *gin.Context) (model.CompanyUpdateInput, string, []byte, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		var input model.CompanyUpdateInput
		if err := c.BindJSON(&input); err != nil {
			return model.CompanyUpdateInput{}, "", nil, errors.New("invalid input body")
		}
		return input, "", nil, nil
	}

	if err := c.Request.ParseMultipartForm(maxAvatarUploadSize + 1024); err != nil {
		return model.CompanyUpdateInput{}, "", nil, errors.New("invalid multipart form")
	}

	input := model.CompanyUpdateInput{}
	if _, ok := c.Request.MultipartForm.Value["name"]; ok {
		value := c.PostForm("name")
		input.Name = &value
	}
	if _, ok := c.Request.MultipartForm.Value["description"]; ok {
		value := c.PostForm("description")
		input.Description = &value
	}
	if _, ok := c.Request.MultipartForm.Value["avatar_url"]; ok {
		value := c.PostForm("avatar_url")
		input.AvatarURL = &value
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return input, "", nil, nil
		}
		return model.CompanyUpdateInput{}, "", nil, errors.New("failed to read avatar file")
	}
	if fileHeader.Size > maxAvatarUploadSize {
		return model.CompanyUpdateInput{}, "", nil, service.ErrAvatarTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return model.CompanyUpdateInput{}, "", nil, errors.New("failed to open avatar file")
	}
	defer file.Close()

	fileData, err := io.ReadAll(io.LimitReader(file, maxAvatarUploadSize+1))
	if err != nil {
		return model.CompanyUpdateInput{}, "", nil, errors.New("failed to read avatar file")
	}
	if len(fileData) > maxAvatarUploadSize {
		return model.CompanyUpdateInput{}, "", nil, service.ErrAvatarTooLarge
	}

	return input, fileHeader.Filename, fileData, nil
}

func (h *Handler) deleteCompany(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	if err := h.services.Company.DeleteCompany(companyID, int64(userID)); err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) leaveCompany(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	var input struct {
		NewOwnerID *int64 `json:"new_owner_id"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.BindJSON(&input); err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid input body")
			return
		}
	}

	if err := h.services.Company.LeaveCompany(companyID, int64(userID), input.NewOwnerID); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) inviteToCompany(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	var input struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	invite, err := h.services.Company.InviteUser(companyID, int64(userID), input.Username)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, invite)
}

func (h *Handler) listInvitations(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	invites, err := h.services.Company.ListInvitations(int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, invites)
}

func (h *Handler) acceptInvitation(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	inviteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid invitation id")
		return
	}

	if err := h.services.Company.AcceptInvitation(inviteID, int64(userID)); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) declineInvitation(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	inviteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid invitation id")
		return
	}

	if err := h.services.Company.DeclineInvitation(inviteID, int64(userID)); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) listCompanyMembers(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	members, err := h.services.Company.ListCompanyMembers(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if members == nil {
		members = []model.CompanyMemberView{}
	}
	c.JSON(http.StatusOK, members)
}

func (h *Handler) removeCompanyMember(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	memberUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.services.Company.RemoveCompanyMember(companyID, int64(userID), memberUserID); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
