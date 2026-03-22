package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(T *testing.T) {
	// Not parallel: mutates package-level vars.

	T.Run("returns unknown when vars are unset", func(t *testing.T) {
		origVersion, origHash, origCTime, origBTime := Version, CommitHash, CommitTime, BuildTime
		Version, CommitHash, CommitTime, BuildTime = "", "", "", ""
		t.Cleanup(func() {
			Version, CommitHash, CommitTime, BuildTime = origVersion, origHash, origCTime, origBTime
		})

		info := Get()
		assert.Equal(t, "unknown", info.Version)
		assert.Equal(t, "unknown", info.CommitHash)
		assert.Equal(t, "unknown", info.CommitTime)
		assert.Equal(t, "unknown", info.BuildTime)
	})

	T.Run("returns set values when vars are populated", func(t *testing.T) {
		origVersion, origHash, origCTime, origBTime := Version, CommitHash, CommitTime, BuildTime
		Version = "v1.2.3"
		CommitHash = "abc123"
		CommitTime = "2026-01-01T00:00:00Z"
		BuildTime = "2026-01-02T00:00:00Z"
		t.Cleanup(func() {
			Version, CommitHash, CommitTime, BuildTime = origVersion, origHash, origCTime, origBTime
		})

		info := Get()
		assert.Equal(t, "v1.2.3", info.Version)
		assert.Equal(t, "abc123", info.CommitHash)
		assert.Equal(t, "2026-01-01T00:00:00Z", info.CommitTime)
		assert.Equal(t, "2026-01-02T00:00:00Z", info.BuildTime)
	})
}
