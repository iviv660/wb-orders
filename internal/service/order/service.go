package order

import (
	"app/internal/cache"
	"app/internal/repository"
)

type Service struct {
	repo  repository.Repository
	cache cache.Cache
}

func New(repo repository.Repository, cache cache.Cache) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}
