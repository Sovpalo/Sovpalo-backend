package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type EventService struct {
	repo             repository.Event
	availabilityRepo repository.Availability
	notifications    *NotificationService
}

func NewEventService(repo repository.Event, availabilityRepo repository.Availability, notifications *NotificationService) *EventService {
	return &EventService{
		repo:             repo,
		availabilityRepo: availabilityRepo,
		notifications:    notifications,
	}
}

func (s *EventService) CreateEvent(userID int64, input model.EventCreateInput, photoFileName string, photoFileData []byte) (int64, error) {
	if input.Title == "" {
		return 0, errors.New("title is required")
	}
	if input.StartTime == nil {
		return 0, errors.New("start_time is required")
	}

	var newPhotoURL string
	if len(photoFileData) > 0 {
		var err error
		newPhotoURL, err = saveEntityAvatarFile("event", userID, photoFileName, photoFileData)
		if err != nil {
			return 0, err
		}
		input.PhotoURL = &newPhotoURL
	}

	event := model.Event{
		CompanyID:   input.CompanyID,
		CreatedBy:   userID,
		Title:       input.Title,
		Description: input.Description,
		PhotoURL:    input.PhotoURL,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		PlaceName:   input.PlaceName,
		PlaceLink:   input.PlaceLink,
	}
	id, err := s.repo.CreateEvent(event)
	if err != nil {
		if newPhotoURL != "" {
			_ = removeAvatarByURL(newPhotoURL)
		}
		return 0, err
	}
	s.dispatchEventPush(id, userID, event, "Назначена встреча", fmt.Sprintf("Новая встреча: %s", event.Title), "event_created")
	return id, nil
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

func (s *EventService) GetCompanyEventFeatures(companyID int64, eventID int64, userID int64, now time.Time) (model.EventFeatures, error) {
	event, err := s.repo.GetEvent(eventID, userID)
	if err != nil {
		return model.EventFeatures{}, err
	}
	if event.CompanyID == nil || *event.CompanyID != companyID {
		return model.EventFeatures{}, errors.New("event not found")
	}

	attendance, err := s.repo.ListCompanyEventAttendance(companyID, eventID, userID)
	if err != nil {
		return model.EventFeatures{}, err
	}

	features := model.EventFeatures{
		ParticipantsCount: len(attendance),
		HasTitle:          strings.TrimSpace(event.Title) != "",
		TitleLength:       utf8.RuneCountInString(event.Title),
		HasAddress:        hasEventAddress(event),
	}

	if event.StartTime != nil {
		features.Hour = event.StartTime.Hour()
		features.DayOfWeek = normalizeWeekday(event.StartTime.Weekday())
		features.IsWeekend = event.StartTime.Weekday() == time.Saturday || event.StartTime.Weekday() == time.Sunday
		features.DaysUntilMeeting = int(event.StartTime.Sub(now).Hours() / 24)
	}
	if event.StartTime != nil && event.EndTime != nil && event.EndTime.After(*event.StartTime) {
		features.DurationMinutes = int64(event.EndTime.Sub(*event.StartTime) / time.Minute)
	}

	freeByUser := map[int64]bool{}
	if event.CompanyID != nil && event.StartTime != nil && s.availabilityRepo != nil {
		rangeEnd := *event.StartTime
		if event.EndTime != nil && event.EndTime.After(*event.StartTime) {
			rangeEnd = *event.EndTime
		}

		availabilities, err := s.availabilityRepo.ListAvailabilityInRange(companyID, *event.StartTime, rangeEnd)
		if err != nil {
			return model.EventFeatures{}, err
		}

		for _, item := range availabilities {
			if availabilityCoversEvent(item, *event.StartTime, event.EndTime) {
				freeByUser[item.UserID] = true
			}
		}
	}

	for _, item := range attendance {
		if freeByUser[item.UserID] {
			features.FreeParticipantsCount++
		}

		switch item.Status {
		case "going":
			features.ConfirmedCount++
		case "not_going":
			features.DeclinedCount++
		default:
			features.PendingCount++
		}
	}

	if features.ParticipantsCount > 0 {
		features.FreeRatio = float64(features.FreeParticipantsCount) / float64(features.ParticipantsCount)
	}

	return features, nil
}

func (s *EventService) UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput, photoFileName string, photoFileData []byte) error {
	if input.Title != nil && *input.Title == "" {
		return errors.New("title cannot be empty")
	}
	if input.Description != nil && *input.Description == "" {
		return errors.New("description cannot be empty")
	}
	if input.PhotoURL != nil && *input.PhotoURL == "" {
		return errors.New("photo_url cannot be empty")
	}
	if input.PlaceName != nil && *input.PlaceName == "" {
		return errors.New("place_name cannot be empty")
	}
	if input.PlaceLink != nil && *input.PlaceLink == "" {
		return errors.New("place_link cannot be empty")
	}

	event, err := s.repo.GetEvent(eventID, userID)
	if err != nil {
		return err
	}

	var newPhotoURL string
	if len(photoFileData) > 0 {
		newPhotoURL, err = saveEntityAvatarFile("event", eventID, photoFileName, photoFileData)
		if err != nil {
			return err
		}
		input.PhotoURL = &newPhotoURL
	}

	if err := s.repo.UpdateEvent(eventID, userID, input); err != nil {
		if newPhotoURL != "" {
			_ = removeAvatarByURL(newPhotoURL)
		}
		return err
	}

	if eventNotificationFieldsChanged(input) {
		updatedEvent := event
		if input.StartTime != nil {
			updatedEvent.StartTime = input.StartTime
		}
		if input.EndTime != nil {
			updatedEvent.EndTime = input.EndTime
		}
		if input.PlaceName != nil {
			updatedEvent.PlaceName = input.PlaceName
		}
		if input.PlaceLink != nil {
			updatedEvent.PlaceLink = input.PlaceLink
		}
		s.dispatchEventPush(eventID, userID, updatedEvent, "Встреча изменена", fmt.Sprintf("Изменились время или место встречи: %s", updatedEvent.Title), "event_updated")
	}

	if newPhotoURL != "" && event.PhotoURL != nil && *event.PhotoURL != newPhotoURL {
		_ = removeAvatarByURL(*event.PhotoURL)
	}

	return nil
}

func (s *EventService) dispatchEventPush(eventID int64, actorID int64, event model.Event, title string, body string, notificationType string) {
	if s.notifications == nil || s.availabilityRepo == nil || event.CompanyID == nil {
		return
	}
	recipients, err := s.availabilityRepo.ListCompanyMemberIDs(*event.CompanyID)
	if err != nil {
		return
	}
	relatedType := "event"
	for _, userID := range recipients {
		if userID == actorID {
			continue
		}
		notification := model.PushNotification{
			UserID:            userID,
			Type:              notificationType,
			Title:             title,
			Message:           body,
			RelatedEntityType: &relatedType,
			RelatedEntityID:   &eventID,
			Data: map[string]string{
				"event_id":   fmt.Sprintf("%d", eventID),
				"company_id": fmt.Sprintf("%d", *event.CompanyID),
			},
		}
		go s.notifications.Dispatch(notification)
	}
}

func eventNotificationFieldsChanged(input model.EventUpdateInput) bool {
	return input.StartTime != nil || input.EndTime != nil || input.PlaceName != nil || input.PlaceLink != nil
}

func (s *EventService) DeleteEvent(eventID int64, userID int64) error {
	return s.repo.DeleteEvent(eventID, userID)
}

func (s *EventService) SetCompanyEventAttendance(companyID int64, eventID int64, userID int64, status string) error {
	switch status {
	case "unknown", "going", "not_going":
		return s.repo.SetCompanyEventAttendance(companyID, eventID, userID, status)
	default:
		return errors.New("invalid status")
	}
}

func (s *EventService) ListCompanyEventAttendance(companyID int64, eventID int64, userID int64) ([]model.EventAttendanceView, error) {
	return s.repo.ListCompanyEventAttendance(companyID, eventID, userID)
}

func normalizeWeekday(day time.Weekday) int {
	return int(day)
}

func hasEventAddress(event model.Event) bool {
	return (event.PlaceName != nil && strings.TrimSpace(*event.PlaceName) != "") ||
		(event.PlaceLink != nil && strings.TrimSpace(*event.PlaceLink) != "")
}

func availabilityCoversEvent(item model.UserAvailability, start time.Time, end *time.Time) bool {
	if item.StartTime.After(start) {
		return false
	}
	if end == nil || !end.After(start) {
		return item.EndTime.After(start) || item.EndTime.Equal(start)
	}
	return item.EndTime.After(*end) || item.EndTime.Equal(*end)
}
