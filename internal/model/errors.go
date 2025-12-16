package model

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrCacheMiss = errors.New("miss cache")
)
