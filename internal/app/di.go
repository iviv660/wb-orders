package app

import (
	"app/internal/adapter"
	kaf "app/internal/adapter/kafka"
	"app/internal/cache"
	cacheobs "app/internal/cache/obs"
	orderCache "app/internal/cache/order"
	"app/internal/closer"
	"app/internal/config"
	"app/internal/repository"
	repoobs "app/internal/repository/obs"
	repo "app/internal/repository/order"
	serviceInter "app/internal/service"
	service "app/internal/service/order"
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type diContainer struct {
	kafkaReader *kafka.Reader
	dlqWriter   *kafka.Writer
	pgxPool     *pgxpool.Pool
	ttl         time.Duration

	consumer adapter.Consumer
	svc      serviceInter.Service
	cache    cache.Cache
	repo     repository.Repository

	worker *kaf.Worker
}

func NewDIContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) Init(ctx context.Context) error {
	if config.AppConfig == nil {
		return errors.New("config.AppConfig is nil: call config.Init() first")
	}

	if err := d.initTTL(); err != nil {
		return err
	}
	if err := d.initPGX(ctx); err != nil {
		return err
	}
	if err := d.initKafka(); err != nil {
		return err
	}
	return nil
}

func (d *diContainer) OrderAdapter(ctx context.Context) (adapter.Consumer, error) {
	_ = ctx
	if d.consumer != nil {
		return d.consumer, nil
	}
	if d.kafkaReader == nil {
		return nil, errors.New("kafka reader is nil: call Init() first")
	}
	d.consumer = kaf.New(d.kafkaReader)
	return d.consumer, nil
}

func (d *diContainer) OrderService(ctx context.Context) (serviceInter.Service, error) {
	if d.svc != nil {
		return d.svc, nil
	}

	r, err := d.OrderRepository(ctx)
	if err != nil {
		return nil, err
	}
	c, err := d.OrderCache(ctx)
	if err != nil {
		return nil, err
	}

	d.svc = service.New(r, c)
	return d.svc, nil
}

func (d *diContainer) OrderRepository(ctx context.Context) (repository.Repository, error) {
	_ = ctx
	if d.repo != nil {
		return d.repo, nil
	}
	if d.pgxPool == nil {
		return nil, errors.New("pgx pool is nil: call Init() first")
	}

	base := repo.New(d.pgxPool)
	d.repo = repoobs.Wrap(base)

	return d.repo, nil
}

func (d *diContainer) OrderCache(ctx context.Context) (cache.Cache, error) {
	_ = ctx
	if d.cache != nil {
		return d.cache, nil
	}
	if d.ttl <= 0 {
		return nil, errors.New("cache ttl is invalid: call Init() first")
	}

	base := orderCache.New(d.ttl)

	interval := time.Minute
	if d.ttl < interval {
		interval = d.ttl
	}
	base.StartWorker(interval)

	d.cache = cacheobs.Wrap(base)

	closer.AddNamed("order-cache", func(ctx context.Context) error {
		base.Close()
		return nil
	})

	return d.cache, nil
}

func (d *diContainer) Worker(ctx context.Context) (*kaf.Worker, error) {
	if d.worker != nil {
		return d.worker, nil
	}

	consumer, err := d.OrderAdapter(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := d.OrderService(ctx)
	if err != nil {
		return nil, err
	}
	if d.dlqWriter == nil {
		return nil, errors.New("dlq writer is nil: call Init() first")
	}

	d.worker = kaf.NewWorker(consumer, svc, d.dlqWriter)
	return d.worker, nil
}

func (d *diContainer) initTTL() error {
	d.ttl = config.AppConfig.Cache.TTL
	if d.ttl <= 0 {
		d.ttl = 5 * time.Minute
	}
	return nil
}

func (d *diContainer) initPGX(ctx context.Context) error {
	log.Printf("[pg] connecting: %s", config.AppConfig.Postgres.DSN)
	pool, err := pgxpool.New(ctx, config.AppConfig.Postgres.DSN)
	if err != nil {
		log.Printf("[pg] connect failed: %v", err)
		return err
	}
	d.pgxPool = pool
	log.Printf("[pg] pool created")

	closer.AddNamed("pgxpool", func(ctx context.Context) error {
		d.pgxPool.Close()
		return nil
	})
	return nil
}

func (d *diContainer) initKafka() error {
	cfg := config.AppConfig.Kafka

	log.Printf("[kafka] brokers=%v topic=%q group=%q dlq=%q",
		cfg.Brokers, cfg.Topic, cfg.GroupID, cfg.DLQTopic,
	)

	if len(cfg.Brokers) == 0 {
		return errors.New("kafka brokers is empty")
	}
	if cfg.Topic == "" {
		return errors.New("kafka topic is empty")
	}
	if cfg.DLQTopic == "" {
		return errors.New("kafka dlq topic is empty")
	}

	d.kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
		Topic:   cfg.Topic,
	})

	d.dlqWriter = &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.DLQTopic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	closer.AddNamed("kafka-reader", func(ctx context.Context) error {
		return d.kafkaReader.Close()
	})
	closer.AddNamed("kafka-dlq-writer", func(ctx context.Context) error {
		return d.dlqWriter.Close()
	})

	return nil
}
