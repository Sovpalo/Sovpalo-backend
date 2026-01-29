package handler

import (
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	health   service.HealthService
	services *service.Service
	repo     repository.Repository
}

func NewHandler(health service.HealthService) *Handler {
	return &Handler{health: health}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", h.healthHandler)




	return router
}

func (h *Handler) healthHandler(c *gin.Context) {
	status, err := h.health.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": status,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}
