package app

import (
	"app/internal/api"
	v1 "app/internal/api/v1"
	"app/internal/config"
	"app/internal/repository"
	repo "app/internal/repository/order"
	"app/internal/service"
	"app/internal/service/order"
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type diContainer struct {
	cfg config.Config

	orderApi api.OrderServer

	orderService service.Service

	orderRepository repository.OrderRepository

	pool *pgxpool.Pool
}

func NewDIContainer() *diContainer {
	return &diContainer{
		cfg: config.Load(),
	}
}

func (d *diContainer) OrderApi(ctx context.Context) api.OrderServer {
	if d.orderApi == nil {
		d.orderApi = v1.NewAPI(d.OrderService(ctx))
	}
	return d.orderApi
}

func (d *diContainer) OrderService(ctx context.Context) service.Service {
	if d.orderService == nil {
		d.orderService = order.NewService(d.OrderRepository(ctx))
	}
	return d.orderService
}

func (d *diContainer) OrderRepository(ctx context.Context) repository.OrderRepository {
	if d.orderRepository == nil {
		d.orderRepository = repo.NewOrderRepository(d.Pool(ctx))
	}
	return d.orderRepository
}

func (d *diContainer) Pool(ctx context.Context) *pgxpool.Pool {
	if d.pool != nil {
		return d.pool
	}

	pool, err := pgxpool.New(ctx, d.cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("failed to create postgres pool: %v", err)
	}

	d.pool = pool
	return d.pool
}
