package databasecfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			ReadConnection: ConnectionDetails{
				Host:     "localhost",
				Username: "root",
				Password: "password",
				Port:     1234,
				Database: "test",
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConnectionDetails_LoadFromURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleURI := "postgres://dbuser:hunter2@pgdatabase:5432/dinner-done-better?sslmode=disable"

		d := &ConnectionDetails{}

		assert.NoError(t, d.LoadFromURL(exampleURI))

		assert.Equal(t, d.Username, "dbuser")
		assert.Equal(t, d.Password, "hunter2")
		assert.Equal(t, d.Host, "pgdatabase")
		assert.Equal(t, d.Database, "dinner-done-better")
		assert.Equal(t, d.DisableSSL, true)
	})

	T.Run("with invalid port", func(t *testing.T) {
		t.Parallel()

		exampleURI := "postgres://dbuser:hunter2@pgdatabase:5432_yo_2345/dinner-done-better?sslmode=disable"

		d := &ConnectionDetails{}

		assert.Error(t, d.LoadFromURL(exampleURI))
	})
}
