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
	router.GET("/health/smtp", h.smtpHealthHandler)
	auth := router.Group("/auth")
	{
		// старт регистрации и отправка кода на email
		auth.POST("/sign-up", h.signUp)
		// подтверждение email-кода и завершение регистрации
		auth.POST("/sign-up/verify", h.verifySignUp)
		// повторная отправка кода подтверждения
		auth.POST("/sign-up/resend", h.resendSignUpCode)
		// старт входа: проверка пароля и отправка кода на email
		auth.POST("/sign-in", h.signIn)
		// подтверждение кода для входа и выдача токена
		auth.POST("/sign-in/verify", h.verifySignIn)
		// повторная отправка кода для входа
		auth.POST("/sign-in/resend", h.resendSignInCode)
		// запуск восстановления пароля
		auth.POST("/password/forgot", h.forgotPassword)
		// подтверждение кода и установка нового пароля
		auth.POST("/password/verify", h.verifyForgotPassword)
		// повторная отправка кода для восстановления пароля
		auth.POST("/password/resend", h.resendForgotPasswordCode)
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

		// приглашение пользователя в компанию (может любой участник), возвращает id приглашения
		companies.POST("/:id/invitations", h.inviteToCompany)
		// получение списка приглашений в компании, которые получил пользователь - возвращает список компаний и id приглашения для каждой из них
		companies.GET("/invitations", h.listInvitations)
		// принять приглашение в компанию (по id приглашения) - добавляет пользователя в компанию и удаляет приглашение
		companies.POST("/invitations/:id/accept", h.acceptInvitation)
		// отклонить приглашение в компанию (по id приглашения) - удаляет приглашение
		companies.POST("/invitations/:id/decline", h.declineInvitation)
		// получить список участников компании
		companies.GET("/:id/members", h.listCompanyMembers)
		// удалить участника компании (только владелец)
		companies.DELETE("/:id/members/:user_id", h.removeCompanyMember)
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
		// POST /companies/:id/events/:event_id/attendance - set attendance (unknown/going/not_going)
		companyEvents.POST("/:event_id/attendance", h.setCompanyEventAttendance)
		// GET /companies/:id/events/:event_id/attendance - list attendance for event
		companyEvents.GET("/:event_id/attendance", h.listCompanyEventAttendance)
		// GET /companies/:id/events/:event_id/attendance/summary - attendance summary
		companyEvents.GET("/:event_id/attendance/summary", h.listCompanyEventAttendanceSummary)
	}

	companyIdeas := router.Group("/companies/:id/ideas", h.userIdentity)
	{
		// POST /companies/:id/ideas - create idea for company
		companyIdeas.POST("", h.createCompanyIdea)
		// GET /companies/:id/ideas - list company ideas
		companyIdeas.GET("", h.listCompanyIdeas)
		// GET /companies/:id/ideas/:idea_id - get company idea by id
		companyIdeas.GET("/:idea_id", h.getCompanyIdea)
		// POST /companies/:id/ideas/:idea_id/like - like idea
		companyIdeas.POST("/:idea_id/like", h.likeCompanyIdea)
		// DELETE /companies/:id/ideas/:idea_id/like - unlike idea
		companyIdeas.DELETE("/:idea_id/like", h.unlikeCompanyIdea)
	}

	availability := router.Group("/companies/:id/availability", h.userIdentity)
	{
		// POST /companies/:id/availability - add availability interval for current user
		availability.POST("", h.createAvailability)
		// GET /companies/:id/availability - list availability of current user in company
		availability.GET("", h.listAvailability)
		// GET /companies/:id/availability/all - list availability of all members in company
		availability.GET("/all", h.listCompanyAvailability)
		// PATCH /companies/:id/availability/:availability_id - update availability interval
		availability.PATCH("/:availability_id", h.updateAvailability)
		// DELETE /companies/:id/availability/:availability_id - delete availability interval
		availability.DELETE("/:availability_id", h.deleteAvailability)
		// POST /companies/:id/availability/intersections - get intersections in range
		availability.POST("/intersections", h.getAvailabilityIntersections)
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

func (h *Handler) smtpHealthHandler(c *gin.Context) {
	status, err := h.health.SMTPStatus(c.Request.Context())
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
