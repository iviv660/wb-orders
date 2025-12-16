package service

import (
	service "app/internal/model"
	"context"
)

type Service interface {
	ProcessOrder(ctx context.Context, order service.Order) error
	Get(ctx context.Context, uuid string) (service.Order, error)
}
