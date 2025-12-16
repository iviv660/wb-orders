package order

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type OrderRepository struct {
	pool Pool
}

func New(pool Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}
