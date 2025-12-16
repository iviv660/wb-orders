package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"app/internal/adapter"
	"app/internal/adapter/converter"
	adapterModel "app/internal/adapter/model"
	"app/internal/logger"
	serviceModel "app/internal/model"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type OrderService interface {
	ProcessOrder(ctx context.Context, order serviceModel.Order) error
}

type Worker struct {
	consumer  adapter.Consumer
	svc       OrderService
	validate  *validator.Validate
	dlqWriter *kafka.Writer

	maxRetries  int
	baseBackoff time.Duration
	maxBackoff  time.Duration
}

func NewWorker(c adapter.Consumer, svc OrderService, dlq *kafka.Writer) *Worker {
	return &Worker{
		consumer:    c,
		svc:         svc,
		validate:    validator.New(),
		dlqWriter:   dlq,
		maxRetries:  5,
		baseBackoff: 200 * time.Millisecond,
		maxBackoff:  5 * time.Second,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	logger.Info(ctx, "kafka worker started")

	err := w.consumer.Read(ctx, func(ctx context.Context, msg kafka.Message) error {
		var dto adapterModel.OrderDTO
		if err := json.Unmarshal(msg.Value, &dto); err != nil {
			logger.Warn(ctx, "bad message: json",
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
				zap.Error(err),
			)

			if dlqErr := sendToDLQ(ctx, w.dlqWriter, msg, err, 0); dlqErr != nil {
				logger.Error(ctx, "dlq write failed (json error)", zap.Error(dlqErr))
				return dlqErr
			}
			return nil
		}

		if err := w.validate.Struct(dto); err != nil {
			logger.Warn(ctx, "bad message: validation",
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
				zap.Error(err),
			)

			if dlqErr := sendToDLQ(ctx, w.dlqWriter, msg, err, 0); dlqErr != nil {
				logger.Error(ctx, "dlq write failed (validation error)", zap.Error(dlqErr))
				return dlqErr
			}
			return nil
		}

		order := converter.OrderDTOToModel(dto)

		var lastErr error
		for attempt := 1; attempt <= w.maxRetries+1; attempt++ {
			if attempt > 1 {
				if err := sleepCtx(ctx, w.backoff(attempt-1)); err != nil {
					return err
				}
			}

			if err := w.svc.ProcessOrder(ctx, order); err == nil {
				logger.Debug(ctx, "order processed",
					zap.String("order_uid", order.OrderUUID),
					zap.String("topic", msg.Topic),
					zap.Int("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
				)
				return nil
			} else {
				lastErr = err
				logger.Warn(ctx, "process failed",
					zap.String("order_uid", order.OrderUUID),
					zap.Int("attempt", attempt),
					zap.Error(err),
				)

				if !isTemporary(err) {
					break
				}
			}
		}

		logger.Error(ctx, "sending to DLQ after retries",
			zap.String("order_uid", order.OrderUUID),
			zap.Error(lastErr),
		)

		if lastErr == nil {
			lastErr = errors.New("processing failed: unknown error")
		}

		if dlqErr := sendToDLQ(ctx, w.dlqWriter, msg, lastErr, w.maxRetries); dlqErr != nil {
			logger.Error(ctx, "dlq write failed (after retries)", zap.Error(dlqErr))
			return dlqErr
		}

		return nil
	})

	if err != nil && ctx.Err() == nil {
		logger.Error(ctx, "kafka worker stopped with error", zap.Error(err))
		return err
	}

	logger.Info(ctx, "kafka worker stopped")
	return err
}

func (w *Worker) backoff(retryAttempt int) time.Duration {
	d := w.baseBackoff * time.Duration(1<<uint(retryAttempt-1))
	if d > w.maxBackoff {
		return w.maxBackoff
	}
	return d
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func isTemporary(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}
