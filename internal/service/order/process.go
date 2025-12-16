package order

import (
	service "app/internal/model"
	"context"
)

func (s *Service) ProcessOrder(ctx context.Context, order service.Order) error {
	if err := s.repo.SetOrder(ctx, order); err != nil {
		return err
	}
	_ = s.cache.Set("order:"+order.OrderUUID, order)
	return nil
}
