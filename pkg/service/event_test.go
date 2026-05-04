package service

import (
	"errors"
	"testing"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
)

type eventRepoStub struct {
	event      model.Event
	attendance []model.EventAttendanceView
	eventErr   error
	attendErr  error
}

func (s eventRepoStub) CreateEvent(event model.Event) (int64, error) {
	return 0, nil
}

func (s eventRepoStub) GetEvent(eventID int64, userID int64) (model.Event, error) {
	if s.eventErr != nil {
		return model.Event{}, s.eventErr
	}
	return s.event, nil
}

func (s eventRepoStub) ListEvents(userID int64) ([]model.Event, error) {
	return nil, nil
}

func (s eventRepoStub) ListCompanyEvents(companyID int64, userID int64) ([]model.Event, error) {
	return nil, nil
}

func (s eventRepoStub) UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput) error {
	return nil
}

func (s eventRepoStub) DeleteEvent(eventID int64, userID int64) error {
	return nil
}

func (s eventRepoStub) SetCompanyEventAttendance(companyID int64, eventID int64, userID int64, status string) error {
	return nil
}

func (s eventRepoStub) ListCompanyEventAttendance(companyID int64, eventID int64, userID int64) ([]model.EventAttendanceView, error) {
	if s.attendErr != nil {
		return nil, s.attendErr
	}
	return s.attendance, nil
}

func TestEventServiceGetCompanyEventFeaturesReturnsCalculatedValues(t *testing.T) {
	start := time.Date(2026, 5, 2, 14, 30, 0, 0, time.UTC)
	end := time.Date(2026, 5, 2, 16, 0, 0, 0, time.UTC)
	companyID := int64(7)
	address := "Office"

	svc := NewEventService(
		eventRepoStub{
			event: model.Event{
				ID:        10,
				CompanyID: &companyID,
				Title:     "Встреча команды",
				StartTime: &start,
				EndTime:   &end,
				PlaceName: &address,
			},
			attendance: []model.EventAttendanceView{
				{UserID: 1, Status: "going"},
				{UserID: 2, Status: "not_going"},
				{UserID: 3, Status: "unknown"},
			},
		},
		availabilityRepoStub{
			availabilities: []model.UserAvailability{
				{UserID: 1, StartTime: start.Add(-time.Hour), EndTime: end},
				{UserID: 3, StartTime: start.Add(-30 * time.Minute), EndTime: end.Add(-time.Minute)},
			},
		},
	)

	features, err := svc.GetCompanyEventFeatures(companyID, 10, 1, time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if features.Hour != 14 {
		t.Fatalf("expected hour 14, got %d", features.Hour)
	}
	if features.DayOfWeek != 6 {
		t.Fatalf("expected saturday as 6, got %d", features.DayOfWeek)
	}
	if !features.IsWeekend {
		t.Fatal("expected weekend flag to be true")
	}
	if features.ParticipantsCount != 3 {
		t.Fatalf("expected 3 participants, got %d", features.ParticipantsCount)
	}
	if features.FreeParticipantsCount != 1 {
		t.Fatalf("expected 1 free participant, got %d", features.FreeParticipantsCount)
	}
	if features.FreeRatio != float64(1)/3 {
		t.Fatalf("expected free ratio 1/3, got %f", features.FreeRatio)
	}
	if features.ConfirmedCount != 1 || features.DeclinedCount != 1 || features.PendingCount != 1 {
		t.Fatalf("unexpected attendance counts: %+v", features)
	}
	if features.DurationMinutes != 90 {
		t.Fatalf("expected 90 minutes, got %d", features.DurationMinutes)
	}
	if features.DaysUntilMeeting != 2 {
		t.Fatalf("expected 2 days until meeting, got %d", features.DaysUntilMeeting)
	}
	if !features.HasTitle {
		t.Fatal("expected has_title to be true")
	}
	if features.TitleLength != len([]rune("Встреча команды")) {
		t.Fatalf("unexpected title length: %d", features.TitleLength)
	}
	if !features.HasAddress {
		t.Fatal("expected has_address to be true")
	}
}

func TestEventServiceGetCompanyEventFeaturesHandlesInstantMeeting(t *testing.T) {
	start := time.Date(2026, 5, 4, 9, 0, 0, 0, time.UTC)
	companyID := int64(3)

	svc := NewEventService(
		eventRepoStub{
			event: model.Event{
				ID:        11,
				CompanyID: &companyID,
				Title:     "Standup",
				StartTime: &start,
			},
			attendance: []model.EventAttendanceView{
				{UserID: 10, Status: "unknown"},
			},
		},
		availabilityRepoStub{
			availabilities: []model.UserAvailability{
				{UserID: 10, StartTime: start.Add(-time.Minute), EndTime: start},
			},
		},
	)

	features, err := svc.GetCompanyEventFeatures(companyID, 11, 10, start.Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if features.DurationMinutes != 0 {
		t.Fatalf("expected zero duration, got %d", features.DurationMinutes)
	}
	if features.FreeParticipantsCount != 1 {
		t.Fatalf("expected user to be free at event start, got %d", features.FreeParticipantsCount)
	}
	if features.PendingCount != 1 {
		t.Fatalf("expected one pending participant, got %d", features.PendingCount)
	}
}

func TestEventServiceGetCompanyEventFeaturesPropagatesRepoError(t *testing.T) {
	expectedErr := errors.New("repo failure")
	svc := NewEventService(
		eventRepoStub{eventErr: expectedErr},
		availabilityRepoStub{},
	)

	_, err := svc.GetCompanyEventFeatures(1, 1, 1, time.Now())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestNormalizeWeekdayReturnsZeroForSunday(t *testing.T) {
	if got := normalizeWeekday(time.Sunday); got != 0 {
		t.Fatalf("expected sunday as 0, got %d", got)
	}
}
