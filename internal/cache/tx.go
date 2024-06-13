package cache

import (
	"context"
	"errors"
	"time"
)

func (c *Cache) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {

	var expiration int64
	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	if key == "" {
		return errors.New("key cannot be empty")
	}

	if value == nil {
		return errors.New("value cannot be nil")
	}

	c.Lock()

	defer c.Unlock()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.items[key]

	if !found {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Value, true
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	c.Lock()
	defer c.Unlock()

	if _, found := c.items[key]; !found {
		return errors.New("Key not found")
	}

	delete(c.items, key)

	return nil
}

func (c *Cache) startGC() {
	go c.gc()
}

func (c *Cache) gc() {
	for {
		time.Sleep(c.cleanupInterval)

		if c.items == nil {
			return
		}

		keys := c.expiredKeys()

		if len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

func (c *Cache) expiredKeys() []string {
	c.RLock()
	var keys []string
	defer c.RUnlock()
	for k, i := range c.items {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}
	return keys

}

func (c *Cache) clearItems(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		delete(c.items, k)
	}
}
