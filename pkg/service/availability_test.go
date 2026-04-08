package service

import (
	"errors"
	"testing"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
)

type availabilityRepoStub struct {
	memberIDs      []int64
	availabilities []model.UserAvailability
	memberErr      error
	rangeErr       error
}

func (s availabilityRepoStub) CreateAvailability(companyID int64, userID int64, input model.AvailabilityCreateInput) (int64, error) {
	return 0, nil
}

func (s availabilityRepoStub) ListAvailability(companyID int64, userID int64) ([]model.UserAvailability, error) {
	return nil, nil
}

func (s availabilityRepoStub) ListCompanyAvailability(companyID int64, userID int64) ([]model.UserAvailability, error) {
	return nil, nil
}

func (s availabilityRepoStub) UpdateAvailability(companyID int64, userID int64, availabilityID int64, input model.AvailabilityCreateInput) error {
	return nil
}

func (s availabilityRepoStub) DeleteAvailability(companyID int64, userID int64, availabilityID int64) error {
	return nil
}

func (s availabilityRepoStub) ListCompanyMemberIDs(companyID int64) ([]int64, error) {
	if s.memberErr != nil {
		return nil, s.memberErr
	}
	return s.memberIDs, nil
}

func (s availabilityRepoStub) ListAvailabilityInRange(companyID int64, start time.Time, end time.Time) ([]model.UserAvailability, error) {
	if s.rangeErr != nil {
		return nil, s.rangeErr
	}
	return s.availabilities, nil
}

func TestAvailabilityServiceGetAvailabilityIntersectionsReturnsIntersection(t *testing.T) {
	svc := NewAvailabilityService(availabilityRepoStub{
		memberIDs: []int64{10, 20},
		availabilities: []model.UserAvailability{
			{
				UserID:    10,
				StartTime: time.Date(2026, 4, 8, 8, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 4, 8, 11, 0, 0, 0, time.UTC),
			},
			{
				UserID:    10,
				StartTime: time.Date(2026, 4, 8, 11, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 4, 8, 13, 0, 0, 0, time.UTC),
			},
			{
				UserID:    20,
				StartTime: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC),
			},
		},
	})

	items, err := svc.GetAvailabilityIntersections(1, 10, model.AvailabilityRangeInput{
		StartTime: time.Date(2026, 4, 8, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 8, 14, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one intersection, got %d", len(items))
	}
	if !items[0].StartTime.Equal(time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start time: %v", items[0].StartTime)
	}
	if !items[0].EndTime.Equal(time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected end time: %v", items[0].EndTime)
	}
}

func TestAvailabilityServiceGetAvailabilityIntersectionsRejectsNonMember(t *testing.T) {
	svc := NewAvailabilityService(availabilityRepoStub{
		memberIDs: []int64{20, 30},
	})

	_, err := svc.GetAvailabilityIntersections(1, 10, model.AvailabilityRangeInput{
		StartTime: time.Date(2026, 4, 8, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
	})
	if err == nil || err.Error() != "user is not a member of the company" {
		t.Fatalf("expected membership error, got %v", err)
	}
}

func TestAvailabilityServiceGetAvailabilityIntersectionsRejectsInvalidRange(t *testing.T) {
	svc := NewAvailabilityService(availabilityRepoStub{})

	_, err := svc.GetAvailabilityIntersections(1, 10, model.AvailabilityRangeInput{
		StartTime: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
	})
	if err == nil || err.Error() != "invalid time range" {
		t.Fatalf("expected invalid range error, got %v", err)
	}
}

func TestAvailabilityServiceGetAvailabilityIntersectionsPropagatesRepoError(t *testing.T) {
	expectedErr := errors.New("repo failure")
	svc := NewAvailabilityService(availabilityRepoStub{
		memberErr: expectedErr,
	})

	_, err := svc.GetAvailabilityIntersections(1, 10, model.AvailabilityRangeInput{
		StartTime: time.Date(2026, 4, 8, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
