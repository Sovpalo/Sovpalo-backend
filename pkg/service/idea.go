package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type IdeaService struct {
	repo        repository.Idea
	companyRepo repository.Company
	generator   IdeaGenerator
}

func NewIdeaService(repo repository.Idea, companyRepo repository.Company, generator IdeaGenerator) *IdeaService {
	return &IdeaService{
		repo:        repo,
		companyRepo: companyRepo,
		generator:   generator,
	}
}

func (s *IdeaService) CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput, photoFileName string, photoFileData []byte) (int64, error) {
	if input.Title == "" {
		return 0, errors.New("title is required")
	}

	var newPhotoURL string
	if len(photoFileData) > 0 {
		var err error
		newPhotoURL, err = saveEntityAvatarFile("idea", userID, photoFileName, photoFileData)
		if err != nil {
			return 0, err
		}
		input.PhotoURL = &newPhotoURL
	}

	id, err := s.repo.CreateCompanyIdea(companyID, userID, input)
	if err != nil {
		if newPhotoURL != "" {
			_ = removeAvatarByURL(newPhotoURL)
		}
		return 0, err
	}
	return id, nil
}

func (s *IdeaService) GenerateCompanyIdeas(ctx context.Context, companyID int64, userID int64, input model.IdeaGenerateInput) ([]model.GeneratedIdeaDraft, error) {
	if strings.TrimSpace(input.Topic) == "" {
		return nil, ErrIdeaTopicRequired
	}

	count := input.Count
	if count == 0 {
		count = defaultIdeaCount
	}
	if count < 1 || count > maxIdeaCount {
		return nil, ErrIdeaCountOutOfRange
	}

	company, err := s.companyRepo.GetCompany(companyID, userID)
	if err != nil {
		return nil, err
	}

	request := IdeaGenerationRequest{
		CompanyName: company.Name,
		Topic:       input.Topic,
		Context:     input.Context,
		Audience:    input.Audience,
		Constraints: input.Constraints,
		Tone:        input.Tone,
		Count:       count,
	}
	prompt := buildIdeaPrompt(request, count)

	items, err := s.generator.GenerateIdeas(ctx, request)
	if err != nil {
		return nil, err
	}

	result := make([]model.GeneratedIdeaDraft, 0, len(items))
	for _, item := range items {
		result = append(result, model.GeneratedIdeaDraft{
			Title:       item.Title,
			Description: item.Description,
			Source:      yandexIdeaSourceName,
			LLMPrompt:   prompt,
		})
	}

	return result, nil
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
