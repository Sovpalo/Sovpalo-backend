package service

import (
	"errors"
	"sort"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type AvailabilityService struct {
	repo repository.Availability
}

func NewAvailabilityService(repo repository.Availability) *AvailabilityService {
	return &AvailabilityService{repo: repo}
}

func (s *AvailabilityService) CreateAvailability(companyID int64, userID int64, input model.AvailabilityCreateInput) (int64, error) {
	if input.EndTime.Before(input.StartTime) || input.EndTime.Equal(input.StartTime) {
		return 0, errors.New("invalid time range")
	}
	return s.repo.CreateAvailability(companyID, userID, input)
}

func (s *AvailabilityService) ListAvailability(companyID int64, userID int64) ([]model.UserAvailability, error) {
	return s.repo.ListAvailability(companyID, userID)
}

func (s *AvailabilityService) ListCompanyAvailability(companyID int64, userID int64) ([]model.UserAvailability, error) {
	return s.repo.ListCompanyAvailability(companyID, userID)
}

func (s *AvailabilityService) UpdateAvailability(companyID int64, userID int64, availabilityID int64, input model.AvailabilityCreateInput) error {
	if input.EndTime.Before(input.StartTime) || input.EndTime.Equal(input.StartTime) {
		return errors.New("invalid time range")
	}
	return s.repo.UpdateAvailability(companyID, userID, availabilityID, input)
}

func (s *AvailabilityService) DeleteAvailability(companyID int64, userID int64, availabilityID int64) error {
	return s.repo.DeleteAvailability(companyID, userID, availabilityID)
}

func (s *AvailabilityService) GetAvailabilityIntersections(companyID int64, userID int64, input model.AvailabilityRangeInput) ([]model.AvailabilityIntersection, error) {
	if input.EndTime.Before(input.StartTime) || input.EndTime.Equal(input.StartTime) {
		return nil, errors.New("invalid time range")
	}

	memberIDs, err := s.repo.ListCompanyMemberIDs(companyID)
	if err != nil {
		return nil, err
	}
	if len(memberIDs) == 0 {
		return nil, errors.New("company has no members")
	}
	if !containsID(memberIDs, userID) {
		return nil, errors.New("user is not a member of the company")
	}

	availabilities, err := s.repo.ListAvailabilityInRange(companyID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}

	perUser := make(map[int64][]timeRange, len(memberIDs))
	for _, a := range availabilities {
		perUser[a.UserID] = append(perUser[a.UserID], timeRange{
			Start: maxTime(a.StartTime, input.StartTime),
			End:   minTime(a.EndTime, input.EndTime),
		})
	}

	for _, userID := range memberIDs {
		perUser[userID] = mergeRanges(perUser[userID])
		if len(perUser[userID]) == 0 {
			return []model.AvailabilityIntersection{}, nil
		}
	}

	var intersection []timeRange
	first := true
	for _, userID := range memberIDs {
		if first {
			intersection = append(intersection, perUser[userID]...)
			first = false
			continue
		}
		intersection = intersectRanges(intersection, perUser[userID])
		if len(intersection) == 0 {
			return []model.AvailabilityIntersection{}, nil
		}
	}

	result := make([]model.AvailabilityIntersection, 0, len(intersection))
	for _, r := range intersection {
		result = append(result, model.AvailabilityIntersection{
			StartTime: r.Start,
			EndTime:   r.End,
		})
	}
	return result, nil
}

type timeRange struct {
	Start time.Time
	End   time.Time
}

func mergeRanges(ranges []timeRange) []timeRange {
	if len(ranges) == 0 {
		return nil
	}
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].Start.Equal(ranges[j].Start) {
			return ranges[i].End.Before(ranges[j].End)
		}
		return ranges[i].Start.Before(ranges[j].Start)
	})

	merged := []timeRange{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if !r.Start.After(last.End) {
			if r.End.After(last.End) {
				last.End = r.End
			}
			continue
		}
		merged = append(merged, r)
	}
	return merged
}

func intersectRanges(a, b []timeRange) []timeRange {
	var result []timeRange
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		start := maxTime(a[i].Start, b[j].Start)
		end := minTime(a[i].End, b[j].End)
		if end.After(start) {
			result = append(result, timeRange{Start: start, End: end})
		}
		if a[i].End.Before(b[j].End) || a[i].End.Equal(b[j].End) {
			i++
		} else {
			j++
		}
	}
	return result
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func containsID(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
