package order

import (
	service "app/internal/model"
	"context"
)

func (s *Service) Get(ctx context.Context, uuid string) (service.Order, error) {
	if v, ok := s.cache[uuid]; ok {
		return v, nil
	}

	order, err := s.repo.GetOrder(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}

	s.cache[uuid] = order
	
	return order, nil
}
