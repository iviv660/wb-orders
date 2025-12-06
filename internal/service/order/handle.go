package order

import (
	service "app/internal/model"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

func (s *Service) HandleMessage(ctx context.Context, msg kafka.Message) error {
	var order service.Order

	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return err
	}

	if err := s.repo.SetOrder(ctx, order); err != nil {
		return err
	}

	s.cache[order.OrderUUID] = order
	return nil
}
