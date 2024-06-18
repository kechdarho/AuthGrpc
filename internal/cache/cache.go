package cache

import (
	"context"
	"time"
)

type Cacher interface {
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, bool)
	Delete(ctx context.Context, key string) error
}
