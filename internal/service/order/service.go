package order

import (
	service "app/internal/model"
	"app/internal/repository"
)

type Service struct {
	repo  repository.OrderRepository
	cache map[string]service.Order
}

func NewService(repo repository.OrderRepository) *Service {
	return &Service{
		repo:  repo,
		cache: make(map[string]service.Order),
	}
}
