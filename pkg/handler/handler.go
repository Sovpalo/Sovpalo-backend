package handler

import (
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	health   service.HealthService
	services *service.Service
}

func NewHandler(health service.HealthService, services *service.Service) *Handler {
	return &Handler{
		health:   health,
		services: services,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// проверка статуса сервиса, возвращает статус и ошибку, если сервис не работает
	router.GET("/health", h.healthHandler)
	auth := router.Group("/auth")
	{
		// регистрация и возвращение токена
		auth.POST("/sign-up", h.signUp)
		// аутентификация и возвращение токена
		auth.POST("/sign-in", h.signIn)
	}

	companies := router.Group("/companies", h.userIdentity)
	{
		// создание компании, возвращает id новой компании
		companies.POST("", h.createCompany)
		// получение списка компаний, в которых состоит пользователь
		companies.GET("", h.listCompanies)
		// получение информации о компании, если пользователь состоит в ней
		companies.GET("/:id", h.getCompany)
		// обновление информации о компании (только владелец может обновлять) - можно менять название, описание и аватар
		companies.PATCH("/:id", h.updateCompany)
		// удаление компании (только владелец может удалять) - удаляет компанию и всех её членов
		companies.DELETE("/:id", h.deleteCompany)

		// приглашение пользователя в компанию, возвращает id приглашения
		companies.POST("/:id/invitations", h.inviteToCompany)
		// получение списка приглашений в компании, которые получил пользователь - возвращает список компаний и id приглашения для каждой из них
		companies.GET("/invitations", h.listInvitations)
		// принять приглашение в компанию (по id приглашения) - добавляет пользователя в компанию и удаляет приглашение
		companies.POST("/invitations/:id/accept", h.acceptInvitation)
		// отклонить приглашение в компанию (по id приглашения) - удаляет приглашение
		companies.POST("/invitations/:id/decline", h.declineInvitation)
	}

	events := router.Group("/events", h.userIdentity)
	{
		// POST /events - create event (title, start_time, optional company_id)
		events.POST("", h.createEvent)
		// GET /events - list events for current user
		events.GET("", h.listEvents)
		// GET /events/:id - get event by id
		events.GET("/:id", h.getEvent)
		// PATCH /events/:id - update event by id
		events.PATCH("/:id", h.updateEvent)
		// DELETE /events/:id - delete event by id
		events.DELETE("/:id", h.deleteEvent)
	}

	companyEvents := router.Group("/companies/:id/events", h.userIdentity)
	{
		// POST /companies/:id/events - create event for company
		companyEvents.POST("", h.createCompanyEvent)
		// GET /companies/:id/events - list company events
		companyEvents.GET("", h.listCompanyEvents)
		// GET /companies/:id/events/:event_id - get company event by id
		companyEvents.GET("/:event_id", h.getCompanyEvent)
		// PATCH /companies/:id/events/:event_id - update company event by id
		companyEvents.PATCH("/:event_id", h.updateCompanyEvent)
		// DELETE /companies/:id/events/:event_id - delete company event by id
		companyEvents.DELETE("/:event_id", h.deleteCompanyEvent)
	}

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
