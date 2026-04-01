package service

import (
	"context"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type HealthService interface {
	Status(ctx context.Context) (string, error)
	SMTPStatus(ctx context.Context) (string, error)
}

type Health struct {
	repo repository.HealthRepository
}

func NewHealthService(repo repository.HealthRepository) *Health {
	return &Health{repo: repo}
}

func (s *Health) Status(ctx context.Context) (string, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return "db_error", err
	}
	return "ok", nil
}

func (s *Health) SMTPStatus(ctx context.Context) (string, error) {
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := smtpAuthCheck(healthCtx); err != nil {
		return "smtp_error", err
	}

	return "ok", nil
}
