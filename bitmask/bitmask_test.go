package bitmask

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPerm uint8

const (
	permRead   testPerm = 1 << iota // 1
	permWrite                       // 2
	permDelete                      // 4
	permAdmin                       // 8
)

func TestNew(T *testing.T) {
	T.Parallel()

	T.Run("with no flags", func(t *testing.T) {
		t.Parallel()

		mask := New[testPerm]()

		assert.Equal(t, testPerm(0), mask.Value())
		assert.True(t, mask.IsEmpty())
	})

	T.Run("with single flag", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.Equal(t, permRead, mask.Value())
	})

	T.Run("with multiple flags", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permWrite)

		assert.Equal(t, permRead|permWrite, mask.Value())
	})

	T.Run("with duplicate flags", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permRead)

		assert.Equal(t, permRead, mask.Value())
	})
}

func TestFromValue(T *testing.T) {
	T.Parallel()

	T.Run("with zero", func(t *testing.T) {
		t.Parallel()

		mask := FromValue(testPerm(0))

		assert.True(t, mask.IsEmpty())
	})

	T.Run("with specific value", func(t *testing.T) {
		t.Parallel()

		mask := FromValue(testPerm(5))

		assert.True(t, mask.Has(permRead))
		assert.True(t, mask.Has(permDelete))
		assert.False(t, mask.Has(permWrite))
	})
}

func TestBitmask_Value(T *testing.T) {
	T.Parallel()

	T.Run("returns underlying value", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permDelete)

		assert.Equal(t, permRead|permDelete, mask.Value())
	})
}

func TestBitmask_Set(T *testing.T) {
	T.Parallel()

	T.Run("set single flag", func(t *testing.T) {
		t.Parallel()

		base := New[testPerm]()
		mask := base.Set(permRead)

		assert.True(t, mask.Has(permRead))
	})

	T.Run("set multiple flags", func(t *testing.T) {
		t.Parallel()

		base := New[testPerm]()
		mask := base.Set(permRead, permWrite)

		assert.True(t, mask.Has(permRead))
		assert.True(t, mask.Has(permWrite))
	})

	T.Run("set already set flag", func(t *testing.T) {
		t.Parallel()

		base := New(permRead)
		mask := base.Set(permRead)

		assert.Equal(t, permRead, mask.Value())
	})

	T.Run("does not mutate original", func(t *testing.T) {
		t.Parallel()

		original := New(permRead)
		_ = original.Set(permWrite)

		assert.False(t, original.Has(permWrite))
	})
}

func TestBitmask_Clear(T *testing.T) {
	T.Parallel()

	T.Run("clear set flag", func(t *testing.T) {
		t.Parallel()

		base := New(permRead, permWrite)
		mask := base.Clear(permWrite)

		assert.True(t, mask.Has(permRead))
		assert.False(t, mask.Has(permWrite))
	})

	T.Run("clear unset flag", func(t *testing.T) {
		t.Parallel()

		base := New(permRead)
		mask := base.Clear(permWrite)

		assert.Equal(t, permRead, mask.Value())
	})

	T.Run("clear multiple flags", func(t *testing.T) {
		t.Parallel()

		base := New(permRead, permWrite, permDelete)
		mask := base.Clear(permRead, permWrite)

		assert.False(t, mask.Has(permRead))
		assert.False(t, mask.Has(permWrite))
		assert.True(t, mask.Has(permDelete))
	})

	T.Run("does not mutate original", func(t *testing.T) {
		t.Parallel()

		original := New(permRead, permWrite)
		_ = original.Clear(permWrite)

		assert.True(t, original.Has(permWrite))
	})
}

func TestBitmask_Toggle(T *testing.T) {
	T.Parallel()

	T.Run("toggle unset flag sets it", func(t *testing.T) {
		t.Parallel()

		base := New[testPerm]()
		mask := base.Toggle(permRead)

		assert.True(t, mask.Has(permRead))
	})

	T.Run("toggle set flag clears it", func(t *testing.T) {
		t.Parallel()

		base := New(permRead)
		mask := base.Toggle(permRead)

		assert.False(t, mask.Has(permRead))
	})

	T.Run("toggle multiple flags", func(t *testing.T) {
		t.Parallel()

		base := New(permRead)
		mask := base.Toggle(permRead, permWrite)

		assert.False(t, mask.Has(permRead))
		assert.True(t, mask.Has(permWrite))
	})

	T.Run("does not mutate original", func(t *testing.T) {
		t.Parallel()

		original := New(permRead)
		_ = original.Toggle(permRead)

		assert.True(t, original.Has(permRead))
	})
}

func TestBitmask_Has(T *testing.T) {
	T.Parallel()

	T.Run("returns true for set flag", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permWrite)

		assert.True(t, mask.Has(permRead))
	})

	T.Run("returns false for unset flag", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.Has(permWrite))
	})

	T.Run("returns false for zero flag", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.Has(0))
	})

	T.Run("with empty bitmask", func(t *testing.T) {
		t.Parallel()

		mask := New[testPerm]()

		assert.False(t, mask.Has(permRead))
	})
}

func TestBitmask_HasAll(T *testing.T) {
	T.Parallel()

	T.Run("returns true when all flags set", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permWrite, permDelete)

		assert.True(t, mask.HasAll(permRead, permWrite))
	})

	T.Run("returns false when one flag missing", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.HasAll(permRead, permWrite))
	})

	T.Run("returns false for empty flags", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.HasAll())
	})

	T.Run("returns false for zero flag", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.HasAll(0))
	})
}

func TestBitmask_HasAny(T *testing.T) {
	T.Parallel()

	T.Run("returns true when one flag set", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.True(t, mask.HasAny(permRead, permWrite))
	})

	T.Run("returns false when no flags set", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.HasAny(permWrite, permDelete))
	})

	T.Run("returns false for empty flags", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.HasAny())
	})
}

func TestBitmask_IsEmpty(T *testing.T) {
	T.Parallel()

	T.Run("returns true for empty bitmask", func(t *testing.T) {
		t.Parallel()

		mask := New[testPerm]()

		assert.True(t, mask.IsEmpty())
	})

	T.Run("returns false for non-empty bitmask", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.False(t, mask.IsEmpty())
	})

	T.Run("returns true for zero value", func(t *testing.T) {
		t.Parallel()

		var mask Bitmask[testPerm]

		assert.True(t, mask.IsEmpty())
	})
}

func TestBitmask_Count(T *testing.T) {
	T.Parallel()

	T.Run("counts zero bits", func(t *testing.T) {
		t.Parallel()

		mask := New[testPerm]()

		assert.Equal(t, 0, mask.Count())
	})

	T.Run("counts one bit", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.Equal(t, 1, mask.Count())
	})

	T.Run("counts multiple bits", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permWrite, permDelete)

		assert.Equal(t, 3, mask.Count())
	})

	T.Run("counts all bits", func(t *testing.T) {
		t.Parallel()

		mask := FromValue(^testPerm(0))

		assert.Equal(t, 8, mask.Count())
	})
}

func TestBitmask_Union(T *testing.T) {
	T.Parallel()

	T.Run("combines two masks", func(t *testing.T) {
		t.Parallel()

		a := New(permRead)
		b := New(permWrite)
		result := a.Union(b)

		assert.True(t, result.Has(permRead))
		assert.True(t, result.Has(permWrite))
	})

	T.Run("union with empty", func(t *testing.T) {
		t.Parallel()

		a := New(permRead)
		b := New[testPerm]()
		result := a.Union(b)

		assert.Equal(t, a.Value(), result.Value())
	})

	T.Run("union with self", func(t *testing.T) {
		t.Parallel()

		a := New(permRead, permWrite)
		result := a.Union(a)

		assert.Equal(t, a.Value(), result.Value())
	})
}

func TestBitmask_Intersect(T *testing.T) {
	T.Parallel()

	T.Run("intersects two masks", func(t *testing.T) {
		t.Parallel()

		a := New(permRead, permWrite)
		b := New(permWrite, permDelete)
		result := a.Intersect(b)

		assert.False(t, result.Has(permRead))
		assert.True(t, result.Has(permWrite))
		assert.False(t, result.Has(permDelete))
	})

	T.Run("intersect with no overlap", func(t *testing.T) {
		t.Parallel()

		a := New(permRead)
		b := New(permWrite)
		result := a.Intersect(b)

		assert.True(t, result.IsEmpty())
	})

	T.Run("intersect with self", func(t *testing.T) {
		t.Parallel()

		a := New(permRead, permWrite)
		result := a.Intersect(a)

		assert.Equal(t, a.Value(), result.Value())
	})
}

func TestBitmask_Difference(T *testing.T) {
	T.Parallel()

	T.Run("removes other's flags", func(t *testing.T) {
		t.Parallel()

		a := New(permRead, permWrite, permDelete)
		b := New(permWrite)
		result := a.Difference(b)

		assert.True(t, result.Has(permRead))
		assert.False(t, result.Has(permWrite))
		assert.True(t, result.Has(permDelete))
	})

	T.Run("difference with no overlap", func(t *testing.T) {
		t.Parallel()

		a := New(permRead)
		b := New(permWrite)
		result := a.Difference(b)

		assert.Equal(t, a.Value(), result.Value())
	})

	T.Run("difference with self", func(t *testing.T) {
		t.Parallel()

		a := New(permRead, permWrite)
		result := a.Difference(a)

		assert.True(t, result.IsEmpty())
	})
}

func TestBitmask_String(T *testing.T) {
	T.Parallel()

	T.Run("empty bitmask", func(t *testing.T) {
		t.Parallel()

		mask := New[testPerm]()

		assert.Equal(t, "00000000", mask.String())
	})

	T.Run("single flag", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead)

		assert.Equal(t, "00000001", mask.String())
	})

	T.Run("multiple flags", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permWrite)

		assert.Equal(t, "00000011", mask.String())
	})

	T.Run("all flags", func(t *testing.T) {
		t.Parallel()

		mask := FromValue(^testPerm(0))

		assert.Equal(t, "11111111", mask.String())
	})
}

func TestBitmask_MarshalJSON(T *testing.T) {
	T.Parallel()

	T.Run("marshals as number", func(t *testing.T) {
		t.Parallel()

		mask := New(permRead, permWrite)
		data, err := json.Marshal(&mask)

		require.NoError(t, err)
		assert.Equal(t, "3", string(data))
	})

	T.Run("marshals zero", func(t *testing.T) {
		t.Parallel()

		mask := New[testPerm]()
		data, err := json.Marshal(&mask)

		require.NoError(t, err)
		assert.Equal(t, "0", string(data))
	})

	T.Run("marshals in struct", func(t *testing.T) {
		t.Parallel()

		type wrapper struct {
			Perms Bitmask[testPerm] `json:"perms"`
		}

		w := wrapper{Perms: New(permRead, permDelete)}
		data, err := json.Marshal(&w)

		require.NoError(t, err)
		assert.Equal(t, `{"perms":5}`, string(data))
	})
}

func TestBitmask_UnmarshalJSON(T *testing.T) {
	T.Parallel()

	T.Run("unmarshals number", func(t *testing.T) {
		t.Parallel()

		var mask Bitmask[testPerm]
		err := json.Unmarshal([]byte("3"), &mask)

		require.NoError(t, err)
		assert.True(t, mask.Has(permRead))
		assert.True(t, mask.Has(permWrite))
	})

	T.Run("unmarshals zero", func(t *testing.T) {
		t.Parallel()

		var mask Bitmask[testPerm]
		err := json.Unmarshal([]byte("0"), &mask)

		require.NoError(t, err)
		assert.True(t, mask.IsEmpty())
	})

	T.Run("returns error for invalid input", func(t *testing.T) {
		t.Parallel()

		var mask Bitmask[testPerm]
		err := json.Unmarshal([]byte(`"not a number"`), &mask)

		assert.Error(t, err)
	})

	T.Run("returns error for negative number", func(t *testing.T) {
		t.Parallel()

		var mask Bitmask[testPerm]
		err := json.Unmarshal([]byte("-1"), &mask)

		assert.Error(t, err)
	})

	T.Run("unmarshals in struct", func(t *testing.T) {
		t.Parallel()

		type wrapper struct {
			Perms Bitmask[testPerm] `json:"perms"`
		}

		var w wrapper
		err := json.Unmarshal([]byte(`{"perms":5}`), &w)

		require.NoError(t, err)
		assert.True(t, w.Perms.Has(permRead))
		assert.True(t, w.Perms.Has(permDelete))
	})

	T.Run("round trip", func(t *testing.T) {
		t.Parallel()

		original := New(permRead, permWrite, permAdmin)
		data, err := json.Marshal(&original)
		require.NoError(t, err)

		var restored Bitmask[testPerm]
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.Value(), restored.Value())
	})
}

func TestBitmask_Immutability(T *testing.T) {
	T.Parallel()

	T.Run("chained operations", func(t *testing.T) {
		t.Parallel()

		a := New(permRead)
		b := a.Set(permWrite)
		c := b.Clear(permRead)

		assert.Equal(t, permRead, a.Value())
		assert.Equal(t, permRead|permWrite, b.Value())
		assert.Equal(t, permWrite, c.Value())
	})
}

func TestBitmask_uint16(T *testing.T) {
	T.Parallel()

	type flag16 uint16

	const (
		f1 flag16 = 1 << iota
		f2
		f3
	)

	T.Run("string has 16 digits", func(t *testing.T) {
		t.Parallel()

		mask := New(f1, f3)

		assert.Equal(t, "0000000000000101", mask.String())
	})

	T.Run("operations work", func(t *testing.T) {
		t.Parallel()

		base := New(f1, f2)
		mask := base.Clear(f1)

		assert.False(t, mask.Has(f1))
		assert.True(t, mask.Has(f2))
	})
}

func TestBitmask_uint32(T *testing.T) {
	T.Parallel()

	type flag32 uint32

	T.Run("count with wider type", func(t *testing.T) {
		t.Parallel()

		mask := FromValue(flag32(0b11110000_00001111))

		assert.Equal(t, 8, mask.Count())
	})

	T.Run("string has 32 digits", func(t *testing.T) {
		t.Parallel()

		mask := New(flag32(1))

		assert.Equal(t, 32, len(mask.String()))
	})
}
