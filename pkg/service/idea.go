package service

import (
	"errors"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type IdeaService struct {
	repo repository.Idea
}

func NewIdeaService(repo repository.Idea) *IdeaService {
	return &IdeaService{repo: repo}
}

func (s *IdeaService) CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput) (int64, error) {
	if input.Title == "" {
		return 0, errors.New("title is required")
	}
	return s.repo.CreateCompanyIdea(companyID, userID, input)
}

func (s *IdeaService) ListCompanyIdeas(companyID int64, userID int64) ([]model.IdeaView, error) {
	return s.repo.ListCompanyIdeas(companyID, userID)
}

func (s *IdeaService) GetCompanyIdea(companyID int64, userID int64, ideaID int64) (model.IdeaView, error) {
	return s.repo.GetCompanyIdea(companyID, userID, ideaID)
}

func (s *IdeaService) UpdateCompanyIdea(companyID int64, userID int64, ideaID int64, input model.IdeaUpdateInput, photoFileName string, photoFileData []byte) error {
	if input.Title != nil && *input.Title == "" {
		return errors.New("title cannot be empty")
	}
	if input.Description != nil && *input.Description == "" {
		return errors.New("description cannot be empty")
	}
	if input.PhotoURL != nil && *input.PhotoURL == "" {
		return errors.New("photo_url cannot be empty")
	}

	idea, err := s.repo.GetCompanyIdea(companyID, userID, ideaID)
	if err != nil {
		return err
	}

	var newPhotoURL string
	if len(photoFileData) > 0 {
		newPhotoURL, err = saveEntityAvatarFile("idea", ideaID, photoFileName, photoFileData)
		if err != nil {
			return err
		}
		input.PhotoURL = &newPhotoURL
	}

	if err := s.repo.UpdateCompanyIdea(companyID, userID, ideaID, input); err != nil {
		if newPhotoURL != "" {
			_ = removeAvatarByURL(newPhotoURL)
		}
		return err
	}

	if newPhotoURL != "" && idea.PhotoURL != nil && *idea.PhotoURL != newPhotoURL {
		_ = removeAvatarByURL(*idea.PhotoURL)
	}

	return nil
}

func (s *IdeaService) LikeCompanyIdea(companyID int64, userID int64, ideaID int64) error {
	return s.repo.LikeCompanyIdea(companyID, userID, ideaID)
}

func (s *IdeaService) UnlikeCompanyIdea(companyID int64, userID int64, ideaID int64) error {
	return s.repo.UnlikeCompanyIdea(companyID, userID, ideaID)
}
