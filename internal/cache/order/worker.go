package order

import (
	"app/internal/logger"
	"context"
	"time"
)

func (c *CacheOrder) StartWorker(interval time.Duration) {
	if c.cancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	t := time.NewTicker(interval)
	c.wg.Add(1)

	logger.Info(context.Background(), "cache janitor started")

	go func() {
		defer c.wg.Done()
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				c.cleanupExpired()
			}
		}
	}()
}

func (c *CacheOrder) cleanupExpired() {
	now := time.Now()

	c.mu.Lock()
	for k, v := range c.cache {
		if now.After(v.expiresAt) {
			delete(c.cache, k)
		}
	}
	c.mu.Unlock()
}
