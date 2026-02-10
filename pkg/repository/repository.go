package repository

import (
	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Authorization
	Company
	Event
}

func NewRepository(pool *pgxpool.Pool, cache *redis.Client) *Repository {
	return &Repository{
		Authorization: NewAuthRepository(pool, cache),
		Company:       NewCompanyRepository(pool),
		Event:         NewEventRepository(pool),
	}
}

type Authorization interface {
	UserExists(email string) (bool, error)
	CreateUser(user model.User) (int, error)
	GetUser(email, password string) (model.User, error)
}

type Company interface {
	CreateCompany(company model.Company) (int64, error)
	GetCompany(companyID int64, userID int64) (model.Company, error)
	ListCompanies(userID int64) ([]model.Company, error)
	UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error
	DeleteCompany(companyID int64, userID int64) error

	CreateInvitation(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error)
	ListInvitations(userID int64) ([]model.CompanyInvitationView, error)
	AcceptInvitation(inviteID int64, userID int64) error
	DeclineInvitation(inviteID int64, userID int64) error
}

type Event interface {
	CreateEvent(event model.Event) (int64, error)
	GetEvent(eventID int64, userID int64) (model.Event, error)
	ListEvents(userID int64) ([]model.Event, error)
	ListCompanyEvents(companyID int64, userID int64) ([]model.Event, error)
	UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput) error
	DeleteEvent(eventID int64, userID int64) error
}
