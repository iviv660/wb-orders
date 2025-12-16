package repository

import (
	service "app/internal/model"
	"context"
)

type Repository interface {
	SetOrder(ctx context.Context, order service.Order) error
	GetOrder(ctx context.Context, uuid string) (service.Order, error)
}
