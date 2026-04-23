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

func (s *CompanyService) CreateCompany(userID int64, name string, description *string, avatarURL *string) (int64, error) {
	if name == "" {
		return 0, errors.New("name is required")
	}
	company := model.Company{
		Name:        name,
		Description: description,
		AvatarURL:   avatarURL,
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

func (s *CompanyService) UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput, avatarFileName string, avatarFileData []byte) error {
	company, err := s.repo.GetCompany(companyID, userID)
	if err != nil {
		return err
	}

	var newAvatarURL string
	if len(avatarFileData) > 0 {
		newAvatarURL, err = saveEntityAvatarFile("company", companyID, avatarFileName, avatarFileData)
		if err != nil {
			return err
		}
		input.AvatarURL = &newAvatarURL
	}

	if err := s.repo.UpdateCompany(companyID, userID, input); err != nil {
		if newAvatarURL != "" {
			_ = removeAvatarByURL(newAvatarURL)
		}
		return err
	}

	if newAvatarURL != "" && company.AvatarURL != nil && *company.AvatarURL != newAvatarURL {
		_ = removeAvatarByURL(*company.AvatarURL)
	}

	return nil
}

func (s *CompanyService) DeleteCompany(companyID int64, userID int64) error {
	return s.repo.DeleteCompany(companyID, userID)
}

func (s *CompanyService) LeaveCompany(companyID int64, userID int64, newOwnerID *int64) error {
	return s.repo.LeaveCompany(companyID, userID, newOwnerID)
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

func (s *CompanyService) ListCompanyMembers(companyID int64, userID int64) ([]model.CompanyMemberView, error) {
	return s.repo.ListCompanyMembers(companyID, userID)
}

func (s *CompanyService) RemoveCompanyMember(companyID int64, ownerID int64, memberUserID int64) error {
	return s.repo.RemoveCompanyMember(companyID, ownerID, memberUserID)
}
