package obs

import (
	"app/internal/logger"
	service "app/internal/model"
	"app/internal/repository"
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Repository struct {
	next   repository.Repository
	tracer trace.Tracer
	dur    metric.Float64Histogram
	errs   metric.Int64Counter
}

func Wrap(next repository.Repository) repository.Repository {
	m := otel.Meter("app/repository")

	dur, err := m.Float64Histogram("repo_duration_ms", metric.WithUnit("ms"))
	if err != nil {
		dur, _ = noop.NewMeterProvider().Meter("noop").Float64Histogram("repo_duration_ms")
	}

	errs, err := m.Int64Counter("repo_errors_total")
	if err != nil {
		errs, _ = noop.NewMeterProvider().Meter("noop").Int64Counter("repo_errors_total")
	}

	return &Repository{
		next:   next,
		tracer: otel.Tracer("app/repository"),
		dur:    dur,
		errs:   errs,
	}
}

func (r *Repository) SetOrder(ctx context.Context, order service.Order) (err error) {
	start := time.Now()

	ctx, span := r.tracer.Start(ctx, "repo.SetOrder",
		trace.WithAttributes(attribute.String("order.uid", order.OrderUUID)),
	)
	defer span.End()

	defer func() {
		r.dur.Record(ctx, float64(time.Since(start).Milliseconds()),
			metric.WithAttributes(attribute.String("op", "SetOrder")),
		)
		if err != nil {
			r.errs.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "SetOrder")))
		}
	}()

	err = r.next.SetOrder(ctx, order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo error")
		logger.Error(ctx, "repo set order failed",
			zap.String("order_uid", order.OrderUUID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (r *Repository) GetOrder(ctx context.Context, uuid string) (order service.Order, err error) {
	start := time.Now()

	ctx, span := r.tracer.Start(ctx, "repo.GetOrder",
		trace.WithAttributes(attribute.String("order.uid", uuid)),
	)
	defer span.End()

	defer func() {
		r.dur.Record(ctx, float64(time.Since(start).Milliseconds()),
			metric.WithAttributes(attribute.String("op", "GetOrder")),
		)
		if err != nil {
			r.errs.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "GetOrder")))
		}
	}()

	order, err = r.next.GetOrder(ctx, uuid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo error")
		logger.Error(ctx, "repo get order failed",
			zap.String("order_uid", uuid),
			zap.Error(err),
		)
		return service.Order{}, err
	}

	return order, nil
}
