package order

import "github.com/jackc/pgx/v5/pgxpool"

type OrderRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}
