package cache

import (
	"app/internal/model"
)

type Cache interface {
	Get(key string) (model.Order, error)
	Set(key string, value model.Order) error
	Delete(key string)
}
