package redis

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	rediscontainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	exampleKey = "example"
	redisImage = "docker.io/redis:7-bullseye"
)

type example struct {
	Name string `json:"name"`
}

func buildContainerBackedRedisConfig(t *testing.T) (config *Config, shutdownFunction func(context.Context) error) {
	t.Helper()

	// Use a dedicated context that won't be cancelled for the container lifecycle
	containerCtx := t.Context()

	redisContainer, err := rediscontainers.Run(containerCtx,
		redisImage,
		rediscontainers.WithLogLevel(rediscontainers.LogLevelNotice),
	)
	if err != nil {
		panic(err)
	}

	// Wait a small amount to ensure container is fully ready
	time.Sleep(100 * time.Millisecond)

	redisAddress, err := redisContainer.ConnectionString(containerCtx)
	must.NoError(t, err)

	cfg := &Config{
		QueueAddresses: []string{
			strings.TrimPrefix(redisAddress, "redis://"),
		},
	}

	shutdownFunc := func(shutdownCtx context.Context) error {
		return redisContainer.Terminate(shutdownCtx)
	}

	return cfg, shutdownFunc
}

func Test_redisCacheImpl_Get(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg, containerShutdown := buildContainerBackedRedisConfig(t)
		defer func() {
			test.NoError(t, containerShutdown(ctx))
		}()
		c, err := NewRedisCache[example](cfg, 0, nil, nil, nil, nil)
		must.NoError(t, err)

		exampleContent := &example{Name: t.Name()}
		test.NoError(t, c.Set(ctx, exampleKey, exampleContent))

		actual, getErr := c.Get(ctx, exampleKey)
		test.Eq(t, exampleContent, actual)
		test.NoError(t, getErr)
	})
}

func Test_redisCacheImpl_Set(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg, containerShutdown := buildContainerBackedRedisConfig(t)
		defer func() {
			test.NoError(t, containerShutdown(ctx))
		}()
		c, err := NewRedisCache[example](cfg, 0, nil, nil, nil, nil)
		must.NoError(t, err)

		exampleContent := &example{Name: t.Name()}
		test.NoError(t, c.Set(ctx, exampleKey, exampleContent))
	})
}

func Test_redisCacheImpl_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg, containerShutdown := buildContainerBackedRedisConfig(t)
		defer func() {
			test.NoError(t, containerShutdown(ctx))
		}()
		c, err := NewRedisCache[example](cfg, 0, nil, nil, nil, nil)
		must.NoError(t, err)

		exampleContent := &example{Name: t.Name()}
		test.NoError(t, c.Set(ctx, exampleKey, exampleContent))

		test.NoError(t, c.Delete(ctx, exampleKey))
	})
}
