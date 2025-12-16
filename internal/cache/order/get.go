package order

import (
	"app/internal/model"
	"time"
)

func (c *CacheOrder) Get(key string) (model.Order, error) {
	now := time.Now()

	c.mu.Lock()
	e, ok := c.cache[key]
	c.mu.Unlock()

	if !ok {
		return model.Order{}, model.ErrNotFound
	}

	if now.After(e.expiresAt) {
		c.mu.Lock()

		if e2, ok2 := c.cache[key]; ok2 && now.After(e2.expiresAt) {
			delete(c.cache, key)
		}
		c.mu.Unlock()
		return model.Order{}, model.ErrCacheMiss
	}

	return e.val, nil
}
