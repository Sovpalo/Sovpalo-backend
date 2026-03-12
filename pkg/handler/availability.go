package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/gin-gonic/gin"
)

type availabilityInput struct {
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Note      *string `json:"note,omitempty"`
}

func (h *Handler) createAvailability(c *gin.Context) {
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

	var input availabilityInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
		return
	}
	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
		return
	}

	id, err := h.services.Availability.CreateAvailability(companyID, int64(userID), model.AvailabilityCreateInput{
		StartTime: startTime,
		EndTime:   endTime,
		Note:      input.Note,
	})
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) listAvailability(c *gin.Context) {
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

	items, err := h.services.Availability.ListAvailability(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) listCompanyAvailability(c *gin.Context) {
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

	items, err := h.services.Availability.ListCompanyAvailability(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) updateAvailability(c *gin.Context) {
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

	availabilityID, err := strconv.ParseInt(c.Param("availability_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid availability id")
		return
	}

	var input availabilityInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
		return
	}
	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
		return
	}

	if err := h.services.Availability.UpdateAvailability(companyID, int64(userID), availabilityID, model.AvailabilityCreateInput{
		StartTime: startTime,
		EndTime:   endTime,
		Note:      input.Note,
	}); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deleteAvailability(c *gin.Context) {
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

	availabilityID, err := strconv.ParseInt(c.Param("availability_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid availability id")
		return
	}

	if err := h.services.Availability.DeleteAvailability(companyID, int64(userID), availabilityID); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) getAvailabilityIntersections(c *gin.Context) {
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

	var input availabilityInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid start_time")
		return
	}
	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid end_time")
		return
	}

	intersections, err := h.services.Availability.GetAvailabilityIntersections(companyID, int64(userID), model.AvailabilityRangeInput{
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, intersections)
}
