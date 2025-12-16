package order

import (
	"testing"
	"time"

	"app/internal/model"

	"github.com/stretchr/testify/require"
)

func TestCacheOrder_SetGet_OK(t *testing.T) {
	c := New(200 * time.Millisecond)

	key := "k1"
	want := model.Order{OrderUUID: "uid-1"}

	require.NoError(t, c.Set(key, want))

	got, err := c.Get(key)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestCacheOrder_Get_NotFound(t *testing.T) {
	c := New(time.Minute)

	_, err := c.Get("missing")
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestCacheOrder_Get_Expired_ReturnsCacheMiss_AndDeletes(t *testing.T) {
	c := New(20 * time.Millisecond)

	key := "k1"
	want := model.Order{OrderUUID: "uid-1"}

	require.NoError(t, c.Set(key, want))

	time.Sleep(35 * time.Millisecond)

	_, err := c.Get(key)
	require.ErrorIs(t, err, model.ErrCacheMiss)

	_, err = c.Get(key)
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestCacheOrder_Delete(t *testing.T) {
	c := New(time.Minute)

	key := "k1"
	want := model.Order{OrderUUID: "uid-1"}

	require.NoError(t, c.Set(key, want))

	c.Delete(key)

	_, err := c.Get(key)
	require.ErrorIs(t, err, model.ErrNotFound)
}
