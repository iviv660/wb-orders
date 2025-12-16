package order

import (
	"app/internal/model"
	"time"
)

func (c *CacheOrder) Set(key string, value model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = entry{
		val:       value,
		expiresAt: time.Now().Add(c.ttl),
	}

	return nil
}
