package order

import (
	"app/internal/logger"
	"app/internal/model"
	"context"
	"sync"
	"time"
)

type entry struct {
	val       model.Order
	expiresAt time.Time
}

type CacheOrder struct {
	mu    sync.Mutex
	cache map[string]entry
	ttl   time.Duration

	cancel func()
	wg     sync.WaitGroup
}

func New(ttl time.Duration) *CacheOrder {
	return &CacheOrder{
		cache: make(map[string]entry),
		ttl:   ttl,
	}
}

func (c *CacheOrder) Close() {
	c.mu.Lock()
	cancel := c.cancel
	c.cancel = nil
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	c.wg.Wait()
	logger.Info(context.Background(), "cache closed")
}
