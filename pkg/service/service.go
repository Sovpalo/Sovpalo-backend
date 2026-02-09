package service

import (
	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type Service struct {
	Authorization
	Company
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Company:       NewCompanyService(repos.Company),
	}
}

type Authorization interface {
	CreateUser(user model.User) (int, error)
	ParseToken(token string) (int, error)
	UserExists(email string) (bool, error)
	SendCodeToEmail(to string, code string) error
	GenerateCode() string
	GenerateToken(email, password string) (string, error)
}

type Company interface {
	CreateCompany(userID int64, name string, description *string) (int64, error)
	GetCompany(companyID int64, userID int64) (model.Company, error)
	ListCompanies(userID int64) ([]model.Company, error)
	UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error
	DeleteCompany(companyID int64, userID int64) error

	InviteUser(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error)
	ListInvitations(userID int64) ([]model.CompanyInvitationView, error)
	AcceptInvitation(inviteID int64, userID int64) error
	DeclineInvitation(inviteID int64, userID int64) error
}
