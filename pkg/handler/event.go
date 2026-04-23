package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

type eventInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
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
		PhotoURL:    input.PhotoURL,
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

	updateInput, photoFileName, photoFileData, err := parseEventUpdateInput(c)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Event.UpdateEvent(eventID, int64(userID), updateInput, photoFileName, photoFileData); err != nil {
		if errors.Is(err, service.ErrAvatarTooLarge) || errors.Is(err, service.ErrAvatarInvalidType) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func parseEventUpdateInput(c *gin.Context) (model.EventUpdateInput, string, []byte, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		var input eventInput
		if err := c.BindJSON(&input); err != nil {
			return model.EventUpdateInput{}, "", nil, errors.New("invalid input body")
		}
		updateInput, err := buildEventUpdateInput(input)
		return updateInput, "", nil, err
	}

	if err := c.Request.ParseMultipartForm(maxAvatarUploadSize + 1024); err != nil {
		return model.EventUpdateInput{}, "", nil, errors.New("invalid multipart form")
	}

	updateInput := model.EventUpdateInput{}
	if _, ok := c.Request.MultipartForm.Value["title"]; ok {
		value := c.PostForm("title")
		updateInput.Title = &value
	}
	if _, ok := c.Request.MultipartForm.Value["description"]; ok {
		value := c.PostForm("description")
		updateInput.Description = &value
	}
	if _, ok := c.Request.MultipartForm.Value["photo_url"]; ok {
		value := c.PostForm("photo_url")
		updateInput.PhotoURL = &value
	}
	if _, ok := c.Request.MultipartForm.Value["company_id"]; ok {
		value := c.PostForm("company_id")
		companyID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return model.EventUpdateInput{}, "", nil, errors.New("invalid company_id")
		}
		updateInput.CompanyID = &companyID
	}
	if _, ok := c.Request.MultipartForm.Value["start_time"]; ok {
		value := c.PostForm("start_time")
		if value != "" {
			startTime, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return model.EventUpdateInput{}, "", nil, errors.New("invalid start_time")
			}
			updateInput.StartTime = &startTime
		}
	}
	if _, ok := c.Request.MultipartForm.Value["end_time"]; ok {
		value := c.PostForm("end_time")
		if value != "" {
			endTime, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return model.EventUpdateInput{}, "", nil, errors.New("invalid end_time")
			}
			updateInput.EndTime = &endTime
		}
	}

	fileName, fileData, err := readMultipartImage(c, "photo")
	if err != nil {
		return model.EventUpdateInput{}, "", nil, err
	}

	return updateInput, fileName, fileData, nil
}

func buildEventUpdateInput(input eventInput) (model.EventUpdateInput, error) {
	var startTime *time.Time
	if input.StartTime != "" {
		parsedStart, err := time.Parse(time.RFC3339, input.StartTime)
		if err != nil {
			return model.EventUpdateInput{}, errors.New("invalid start_time")
		}
		startTime = &parsedStart
	}

	var endTime *time.Time
	if input.EndTime != nil && *input.EndTime != "" {
		parsedEnd, err := time.Parse(time.RFC3339, *input.EndTime)
		if err != nil {
			return model.EventUpdateInput{}, errors.New("invalid end_time")
		}
		endTime = &parsedEnd
	}

	updateInput := model.EventUpdateInput{
		CompanyID:   input.CompanyID,
		Description: input.Description,
		PhotoURL:    input.PhotoURL,
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

	return updateInput, nil
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

	if events == nil {
		events = []model.Event{}
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
		PhotoURL:    input.PhotoURL,
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

	updateInput, photoFileName, photoFileData, err := parseEventUpdateInput(c)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if updateInput.CompanyID != nil && *updateInput.CompanyID != companyID {
		newErrorResponse(c, http.StatusBadRequest, "company_id mismatch")
		return
	}
	updateInput.CompanyID = nil

	if err := h.services.Event.UpdateEvent(eventID, int64(userID), updateInput, photoFileName, photoFileData); err != nil {
		if errors.Is(err, service.ErrAvatarTooLarge) || errors.Is(err, service.ErrAvatarInvalidType) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
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

func (h *Handler) setCompanyEventAttendance(c *gin.Context) {
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

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if err := h.services.Event.SetCompanyEventAttendance(companyID, eventID, int64(userID), input.Status); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) listCompanyEventAttendance(c *gin.Context) {
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

	attendance, err := h.services.Event.ListCompanyEventAttendance(companyID, eventID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, attendance)
}

func (h *Handler) listCompanyEventAttendanceSummary(c *gin.Context) {
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

	attendance, err := h.services.Event.ListCompanyEventAttendance(companyID, eventID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	type summary struct {
		Going    []string `json:"going"`
		NotGoing []string `json:"not_going"`
		Unknown  []string `json:"unknown"`
	}

	result := summary{}
	currentStatus := "unknown"
	for _, item := range attendance {
		switch item.Status {
		case "going":
			result.Going = append(result.Going, item.Username)
		case "not_going":
			result.NotGoing = append(result.NotGoing, item.Username)
		default:
			result.Unknown = append(result.Unknown, item.Username)
		}
		if item.UserID == int64(userID) {
			currentStatus = item.Status
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"going":               result.Going,
		"not_going":           result.NotGoing,
		"unknown":             result.Unknown,
		"current_user_status": currentStatus,
	})
}
