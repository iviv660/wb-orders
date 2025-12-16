package kafka

import (
	"app/internal/otelx"
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("app/kafka/consumer")

func (c *Consumer) Read(ctx context.Context, handle func(ctx context.Context, msg kafka.Message) error) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		msgCtx := otelx.ExtractKafka(ctx, &msg)

		msgCtx, span := tracer.Start(
			msgCtx,
			"kafka.consume",
			trace.WithAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination", msg.Topic),
				attribute.Int("messaging.kafka.partition", msg.Partition),
				attribute.Int64("messaging.kafka.offset", msg.Offset),
			),
		)

		err = handle(msgCtx, msg)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "handler error")
			span.End()
			return err
		}

		span.SetStatus(codes.Ok, "ok")
		span.End()

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
