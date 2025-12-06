package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// MessageHandler описывает обработчик одного сообщения Kafka.
// На уровне адаптера мы работаем с "сырым" kafka.Message,
// а внутри HandleMessage ты можешь маппить его на доменные структуры и вызывать сервисы.
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg kafka.Message) error
}

// NewKafkaReader создаёт настроенный kafka.Reader.
func NewKafkaReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1e3,
		MaxBytes: 10e6,
	})
}

// RunConsumer Запуск Consumer-а.
func RunConsumer(ctx context.Context, r *kafka.Reader, handler MessageHandler) error {
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("Ошибка четения сообщения: %v \n", err)
			continue
		}

		if err = handler.HandleMessage(ctx, m); err != nil {
			log.Printf("Ошибка обработки сообщения из kafka (topic=%s, partition=%d, offset=%d): %v \n",
				m.Topic, m.Partition, m.Offset, err)
		}
	}
}
