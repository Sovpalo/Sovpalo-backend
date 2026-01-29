package service

import "github.com/Sovpalo/sovpalo-backend/pkg/repository"

type Service struct {
}

func NewService(repos *repository.Repository) *Service {
	return &Service{}
}
