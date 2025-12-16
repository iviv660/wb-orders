package kafka

import (
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func New(reader *kafka.Reader) *Consumer {
	c := &Consumer{
		reader: reader,
	}
	return c
}
