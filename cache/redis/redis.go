package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/cache"

	"github.com/go-redis/redis/v8"
)

type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type redisCacheImpl[T any] struct {
	client     redisClient
	expiration time.Duration
}

// NewRedisCache builds a new redis-backed cache.
func NewRedisCache[T any](cfg *Config, expiration time.Duration) cache.Cache[T] {
	return &redisCacheImpl[T]{
		client:     buildRedisClient(cfg),
		expiration: expiration,
	}
}

func (i *redisCacheImpl[T]) Get(ctx context.Context, key string) (*T, error) {
	res, err := i.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	b := strings.NewReader(res)

	var x *T
	if err = gob.NewDecoder(b).Decode(&x); err != nil {
		return nil, fmt.Errorf("decoding from cache: %w", err)
	}

	if x == nil {
		return nil, cache.ErrNotFound
	}

	return x, nil
}

func (i *redisCacheImpl[T]) Set(ctx context.Context, key string, value *T) error {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(value); err != nil {
		return fmt.Errorf("encoding for cache: %w", err)
	}

	if setErr := i.client.Set(ctx, key, b.String(), i.expiration).Err(); setErr != nil {
		return setErr
	}

	return nil
}

func (i *redisCacheImpl[T]) Delete(ctx context.Context, key string) error {
	return i.client.Del(ctx, key).Err()
}

// buildRedisClient returns a PublisherProvider for a given address.
func buildRedisClient(cfg *Config) redisClient {
	var c redisClient
	if len(cfg.QueueAddresses) > 1 {
		c = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.QueueAddresses,
			Username:     cfg.Username,
			Password:     cfg.Password,
			DialTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		})
	} else if len(cfg.QueueAddresses) == 1 {
		c = redis.NewClient(&redis.Options{
			Addr:         cfg.QueueAddresses[0],
			Username:     cfg.Username,
			Password:     cfg.Password,
			DialTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		})
	}

	return c
}
