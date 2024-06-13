package cache

import (
	"context"
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	items             map[string]Item
}

type Item struct {
	Value      interface{}
	Created    time.Time
	Expiration int64
}

func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]Item)

	cache := &Cache{
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		items:             items,
	}

	if cleanupInterval > 0 {
		cache.startGC()
	}

	return cache
}

func InitCache(ctx context.Context, DefaultExpiration time.Duration, CleanupInterval time.Duration) (*Cache, error) {
	cache := NewCache(DefaultExpiration, CleanupInterval)
	return cache, nil
}
