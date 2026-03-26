package handler

import (
	"net/http"
	"strconv"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
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
