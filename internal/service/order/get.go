package order

import (
	service "app/internal/model"
	"context"
)

func (s *Service) Get(ctx context.Context, uuid string) (service.Order, error) {
	key := "order:" + uuid

	if order, err := s.cache.Get(key); err == nil {
		return order, nil
	}

	order, err := s.repo.GetOrder(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}

	_ = s.cache.Set(key, order)
	return order, nil
}
