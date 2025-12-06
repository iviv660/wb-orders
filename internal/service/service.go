package service

import (
	service "app/internal/model"
	"context"

	"github.com/segmentio/kafka-go"
)

type Service interface {
	HandleMessage(ctx context.Context, msg kafka.Message) error
	Get(ctx context.Context, uuid string) (service.Order, error)
}
