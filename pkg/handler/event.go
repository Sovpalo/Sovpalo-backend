package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/gin-gonic/gin"
)

type eventInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	StartTime   string  `json:"start_time"`
	EndTime     *string `json:"end_time,omitempty"`
	CompanyID   *int64  `json:"company_id,omitempty"`
}

func (h *Handler) createEvent(c *gin.Context) {
	var input eventInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
		return
	}

	var endTime *time.Time
	if input.EndTime != nil && *input.EndTime != "" {
		parsedEnd, err := time.Parse(time.RFC3339, *input.EndTime)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
			return
		}
		endTime = &parsedEnd
	}

	eventID, err := h.services.Event.CreateEvent(int64(userID), model.EventCreateInput{
		Title:       input.Title,
		Description: input.Description,
		StartTime:   &startTime,
		EndTime:     endTime,
		CompanyID:   input.CompanyID,
	})
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": eventID})
}

func (h *Handler) listEvents(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	events, err := h.services.Event.ListEvents(int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *Handler) getEvent(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.services.Event.GetEvent(eventID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *Handler) updateEvent(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid event id")
		return
	}

	var input eventInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	var startTime *time.Time
	if input.StartTime != "" {
		parsedStart, err := time.Parse(time.RFC3339, input.StartTime)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
			return
		}
		startTime = &parsedStart
	}

	var endTime *time.Time
	if input.EndTime != nil && *input.EndTime != "" {
		parsedEnd, err := time.Parse(time.RFC3339, *input.EndTime)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
			return
		}
		endTime = &parsedEnd
	}

	updateInput := model.EventUpdateInput{
		CompanyID:   input.CompanyID,
		Description: input.Description,
	}
	if input.Title != "" {
		updateInput.Title = &input.Title
	}
	if startTime != nil {
		updateInput.StartTime = startTime
	}
	if endTime != nil {
		updateInput.EndTime = endTime
	}

	if err := h.services.Event.UpdateEvent(eventID, int64(userID), updateInput); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deleteEvent(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid event id")
		return
	}

	if err := h.services.Event.DeleteEvent(eventID, int64(userID)); err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) listCompanyEvents(c *gin.Context) {
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

	events, err := h.services.Event.ListCompanyEvents(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *Handler) createCompanyEvent(c *gin.Context) {
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

	var input eventInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if input.CompanyID != nil && *input.CompanyID != companyID {
		newErrorResponse(c, http.StatusBadRequest, "company_id mismatch")
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
		return
	}

	var endTime *time.Time
	if input.EndTime != nil && *input.EndTime != "" {
		parsedEnd, err := time.Parse(time.RFC3339, *input.EndTime)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
			return
		}
		endTime = &parsedEnd
	}

	eventID, err := h.services.Event.CreateEvent(int64(userID), model.EventCreateInput{
		Title:       input.Title,
		Description: input.Description,
		StartTime:   &startTime,
		EndTime:     endTime,
		CompanyID:   &companyID,
	})
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": eventID})
}

func (h *Handler) getCompanyEvent(c *gin.Context) {
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

	eventID, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.services.Event.GetEvent(eventID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	if event.CompanyID == nil || *event.CompanyID != companyID {
		newErrorResponse(c, http.StatusNotFound, "event not found")
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *Handler) updateCompanyEvent(c *gin.Context) {
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

	eventID, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.services.Event.GetEvent(eventID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	if event.CompanyID == nil || *event.CompanyID != companyID {
		newErrorResponse(c, http.StatusNotFound, "event not found")
		return
	}

	var input eventInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if input.CompanyID != nil && *input.CompanyID != companyID {
		newErrorResponse(c, http.StatusBadRequest, "company_id mismatch")
		return
	}

	var startTime *time.Time
	if input.StartTime != "" {
		parsedStart, err := time.Parse(time.RFC3339, input.StartTime)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
			return
		}
		startTime = &parsedStart
	}

	var endTime *time.Time
	if input.EndTime != nil && *input.EndTime != "" {
		parsedEnd, err := time.Parse(time.RFC3339, *input.EndTime)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
			return
		}
		endTime = &parsedEnd
	}

	updateInput := model.EventUpdateInput{
		CompanyID:   nil,
		Description: input.Description,
	}
	if input.Title != "" {
		updateInput.Title = &input.Title
	}
	if startTime != nil {
		updateInput.StartTime = startTime
	}
	if endTime != nil {
		updateInput.EndTime = endTime
	}

	if err := h.services.Event.UpdateEvent(eventID, int64(userID), updateInput); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deleteCompanyEvent(c *gin.Context) {
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

	eventID, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.services.Event.GetEvent(eventID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	if event.CompanyID == nil || *event.CompanyID != companyID {
		newErrorResponse(c, http.StatusNotFound, "event not found")
		return
	}

	if err := h.services.Event.DeleteEvent(eventID, int64(userID)); err != nil {
		newErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
