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
