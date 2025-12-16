package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

type Consumer struct {
	reader Reader
}

func New(reader Reader) *Consumer {
	return &Consumer{reader: reader}
}
