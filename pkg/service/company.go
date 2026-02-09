package service

import (
	"errors"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type CompanyService struct {
	repo repository.Company
}

func NewCompanyService(repo repository.Company) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) CreateCompany(userID int64, name string, description *string) (int64, error) {
	if name == "" {
		return 0, errors.New("name is required")
	}
	company := model.Company{
		Name:        name,
		Description: description,
		CreatedBy:   userID,
	}
	return s.repo.CreateCompany(company)
}

func (s *CompanyService) GetCompany(companyID int64, userID int64) (model.Company, error) {
	return s.repo.GetCompany(companyID, userID)
}

func (s *CompanyService) ListCompanies(userID int64) ([]model.Company, error) {
	return s.repo.ListCompanies(userID)
}

func (s *CompanyService) UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error {
	return s.repo.UpdateCompany(companyID, userID, input)
}

func (s *CompanyService) DeleteCompany(companyID int64, userID int64) error {
	return s.repo.DeleteCompany(companyID, userID)
}

func (s *CompanyService) InviteUser(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error) {
	if username == "" {
		return model.CompanyInvitation{}, errors.New("username is required")
	}
	return s.repo.CreateInvitation(companyID, invitedBy, username)
}

func (s *CompanyService) ListInvitations(userID int64) ([]model.CompanyInvitationView, error) {
	return s.repo.ListInvitations(userID)
}

func (s *CompanyService) AcceptInvitation(inviteID int64, userID int64) error {
	return s.repo.AcceptInvitation(inviteID, userID)
}

func (s *CompanyService) DeclineInvitation(inviteID int64, userID int64) error {
	return s.repo.DeclineInvitation(inviteID, userID)
}
