package obs

import (
	"app/internal/cache"
	"app/internal/logger"
	"app/internal/model"
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

type Cache struct {
	next   cache.Cache
	tracer trace.Tracer

	dur  metric.Float64Histogram
	errs metric.Int64Counter
	hits metric.Int64Counter
	miss metric.Int64Counter
}

func Wrap(next cache.Cache) cache.Cache {
	m := otel.Meter("app/cache")

	dur, err := m.Float64Histogram("cache_duration_ms", metric.WithUnit("ms"))
	if err != nil {
		dur, _ = noop.NewMeterProvider().Meter("noop").Float64Histogram("cache_duration_ms")
	}

	errs, err := m.Int64Counter("cache_errors_total")
	if err != nil {
		errs, _ = noop.NewMeterProvider().Meter("noop").Int64Counter("cache_errors_total")
	}

	hits, err := m.Int64Counter("cache_hits_total")
	if err != nil {
		hits, _ = noop.NewMeterProvider().Meter("noop").Int64Counter("cache_hits_total")
	}

	miss, err := m.Int64Counter("cache_miss_total")
	if err != nil {
		miss, _ = noop.NewMeterProvider().Meter("noop").Int64Counter("cache_miss_total")
	}

	return &Cache{
		next:   next,
		tracer: otel.Tracer("app/cache"),
		dur:    dur,
		errs:   errs,
		hits:   hits,
		miss:   miss,
	}
}

func (c *Cache) Get(key string) (model.Order, error) {
	ctx := context.Background()
	start := time.Now()

	ctx, span := c.tracer.Start(ctx, "cache.Get",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	v, err := c.next.Get(key)

	c.dur.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(attribute.String("op", "Get")),
	)

	if err == nil {
		c.hits.Add(ctx, 1)
		span.SetStatus(codes.Ok, "hit")
		return v, nil
	}

	c.miss.Add(ctx, 1)
	span.RecordError(err)
	span.SetStatus(codes.Error, "miss")

	if err != model.ErrNotFound && err != model.ErrCacheMiss {
		c.errs.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "Get")))
		logger.Warn(ctx, "cache get failed", zap.String("key", key), zap.Error(err))
	}

	return model.Order{}, err
}

func (c *Cache) Set(key string, value model.Order) error {
	ctx := context.Background()
	start := time.Now()

	ctx, span := c.tracer.Start(ctx, "cache.Set",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	err := c.next.Set(key, value)

	c.dur.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(attribute.String("op", "Set")),
	)

	if err != nil {
		c.errs.Add(ctx, 1, metric.WithAttributes(attribute.String("op", "Set")))
		span.RecordError(err)
		span.SetStatus(codes.Error, "error")
		logger.Warn(ctx, "cache set failed", zap.String("key", key), zap.Error(err))
		return err
	}

	span.SetStatus(codes.Ok, "ok")
	return nil
}

func (c *Cache) Delete(key string) {
	ctx := context.Background()
	start := time.Now()

	ctx, span := c.tracer.Start(ctx, "cache.Delete",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	c.next.Delete(key)

	c.dur.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(attribute.String("op", "Delete")),
	)

	span.SetStatus(codes.Ok, "ok")
}
