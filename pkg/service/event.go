package service

import (
	"errors"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type EventService struct {
	repo repository.Event
}

func NewEventService(repo repository.Event) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) CreateEvent(userID int64, input model.EventCreateInput) (int64, error) {
	if input.Title == "" {
		return 0, errors.New("title is required")
	}
	if input.StartTime == nil {
		return 0, errors.New("start_time is required")
	}
	event := model.Event{
		CompanyID:   input.CompanyID,
		CreatedBy:   userID,
		Title:       input.Title,
		Description: input.Description,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
	}
	return s.repo.CreateEvent(event)
}

func (s *EventService) GetEvent(eventID int64, userID int64) (model.Event, error) {
	return s.repo.GetEvent(eventID, userID)
}

func (s *EventService) ListEvents(userID int64) ([]model.Event, error) {
	return s.repo.ListEvents(userID)
}

func (s *EventService) ListCompanyEvents(companyID int64, userID int64) ([]model.Event, error) {
	return s.repo.ListCompanyEvents(companyID, userID)
}

func (s *EventService) UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput) error {
	if input.Title != nil && *input.Title == "" {
		return errors.New("title cannot be empty")
	}
	if input.Description != nil && *input.Description == "" {
		return errors.New("description cannot be empty")
	}
	return s.repo.UpdateEvent(eventID, userID, input)
}

func (s *EventService) DeleteEvent(eventID int64, userID int64) error {
	return s.repo.DeleteEvent(eventID, userID)
}
