package elasticsearch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(T *testing.T) {
	T.Parallel()

	T.Run("zero value", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		assert.Empty(t, cfg.Address)
		assert.Empty(t, cfg.Username)
		assert.Empty(t, cfg.Password)
		assert.Nil(t, cfg.CACert)
		assert.Equal(t, time.Duration(0), cfg.IndexOperationTimeout)
	})

	T.Run("with values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Address:               "http://localhost:9200",
			Username:              "elastic",
			Password:              "password",
			CACert:                []byte("cert"),
			IndexOperationTimeout: 5 * time.Second,
		}

		assert.Equal(t, "http://localhost:9200", cfg.Address)
		assert.Equal(t, "elastic", cfg.Username)
		assert.Equal(t, "password", cfg.Password)
		assert.Equal(t, []byte("cert"), cfg.CACert)
		assert.Equal(t, 5*time.Second, cfg.IndexOperationTimeout)
	})
}
