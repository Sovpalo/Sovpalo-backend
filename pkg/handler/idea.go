package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createCompanyIdea(c *gin.Context) {
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

	var input model.IdeaCreateInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	id, err := h.services.Idea.CreateCompanyIdea(companyID, int64(userID), input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) listCompanyIdeas(c *gin.Context) {
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

	ideas, err := h.services.Idea.ListCompanyIdeas(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if ideas == nil {
		ideas = []model.IdeaView{}
	}
	c.JSON(http.StatusOK, ideas)
}

func (h *Handler) getCompanyIdea(c *gin.Context) {
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

	ideaID, err := strconv.ParseInt(c.Param("idea_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid idea id")
		return
	}

	idea, err := h.services.Idea.GetCompanyIdea(companyID, int64(userID), ideaID)
	if err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, idea)
}

func (h *Handler) updateCompanyIdea(c *gin.Context) {
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

	ideaID, err := strconv.ParseInt(c.Param("idea_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid idea id")
		return
	}

	input, photoFileName, photoFileData, err := parseIdeaUpdateInput(c)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Idea.UpdateCompanyIdea(companyID, int64(userID), ideaID, input, photoFileName, photoFileData); err != nil {
		if errors.Is(err, service.ErrAvatarTooLarge) || errors.Is(err, service.ErrAvatarInvalidType) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func parseIdeaUpdateInput(c *gin.Context) (model.IdeaUpdateInput, string, []byte, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		var input model.IdeaUpdateInput
		if err := c.BindJSON(&input); err != nil {
			return model.IdeaUpdateInput{}, "", nil, errors.New("invalid input body")
		}
		return input, "", nil, nil
	}

	if err := c.Request.ParseMultipartForm(maxAvatarUploadSize + 1024); err != nil {
		return model.IdeaUpdateInput{}, "", nil, errors.New("invalid multipart form")
	}

	input := model.IdeaUpdateInput{}
	if _, ok := c.Request.MultipartForm.Value["title"]; ok {
		value := c.PostForm("title")
		input.Title = &value
	}
	if _, ok := c.Request.MultipartForm.Value["description"]; ok {
		value := c.PostForm("description")
		input.Description = &value
	}
	if _, ok := c.Request.MultipartForm.Value["photo_url"]; ok {
		value := c.PostForm("photo_url")
		input.PhotoURL = &value
	}

	fileName, fileData, err := readMultipartImage(c, "photo")
	if err != nil {
		return model.IdeaUpdateInput{}, "", nil, err
	}

	return input, fileName, fileData, nil
}

func (h *Handler) likeCompanyIdea(c *gin.Context) {
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

	ideaID, err := strconv.ParseInt(c.Param("idea_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid idea id")
		return
	}

	if err := h.services.Idea.LikeCompanyIdea(companyID, int64(userID), ideaID); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) unlikeCompanyIdea(c *gin.Context) {
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

	ideaID, err := strconv.ParseInt(c.Param("idea_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid idea id")
		return
	}

	if err := h.services.Idea.UnlikeCompanyIdea(companyID, int64(userID), ideaID); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
