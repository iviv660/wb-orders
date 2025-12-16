package adapter

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Read(ctx context.Context, handle func(ctx context.Context, msg kafka.Message) error) error
}
