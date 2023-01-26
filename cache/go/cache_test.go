package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func _contextWithCancel() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, func() {
		cancel()
		time.Sleep(shutdownWindow + 10*time.Millisecond)
	}
}

func _new(ctx context.Context, updateExpiry time.Duration) *cache {
	c := &cache{
		lookup:       make(map[interface{}]*value),
		updateExpiry: updateExpiry,
		requestCh:    make(chan *request),
		gcInterval:   10 * time.Millisecond,
	}

	go c.serve()
	go c.exit(ctx)
	go c.gc(ctx)
	return c
}

func TestCache_Get(t *testing.T) {
	t.Run("key found", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()

		c := _new(ctx, 0)
		c.lookup["k"] = &value{
			updateTime: time.Now(),
			data:       "v",
		}

		val, err := c.Get("k")
		assert.Equal(t, "v", val)
		assert.NoError(t, err)
	})

	t.Run("key not found", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()
		c := _new(ctx, 0)
		_, err := c.Get("k")
		assert.Equal(t, ErrNotFound, err)
	})
}

func TestCache_Set(t *testing.T) {
	t.Run("key not existed", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()

		c := _new(ctx, 0)
		val, err := c.Set("k", "v")
		assert.NoError(t, err)
		assert.Nil(t, val)

		val, err = c.Get("k")
		assert.NoError(t, err)
		assert.Equal(t, "v", val)
	})

	t.Run("key existed", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()

		c := _new(ctx, 0)
		val, err := c.Set("k", "v1")
		assert.NoError(t, err)
		assert.Nil(t, val)

		preVal, err := c.Set("k", "v2")
		assert.NoError(t, err)
		assert.Equal(t, "v1", preVal)
	})
}

func TestCache_Del(t *testing.T) {
	t.Run("key existed", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()

		c := _new(ctx, 0)
		val, err := c.Set("k", "v")
		assert.NoError(t, err)
		assert.Nil(t, val)

		err = c.Del("k")
		assert.NoError(t, err)

		_, err = c.Get("k")
		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("key not existed", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()

		c := _new(ctx, 0)
		err := c.Del("k")
		assert.Equal(t, ErrNotFound, err)
	})
}

func TestCache_Exp(t *testing.T) {
	t.Run("expired key should be garbage collected", func(t *testing.T) {
		ctx, cancel := _contextWithCancel()
		defer cancel()

		c := _new(ctx, 10*time.Millisecond)
		_, err := c.Set("k", "v")
		assert.NoError(t, err)
		time.Sleep(20 * time.Millisecond)
		_, err = c.Get("k")
		assert.Equal(t, ErrNotFound, err)
	})
}

func TestCache_shutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := _new(ctx, 0)
	cancel()
	time.Sleep(10 * time.Millisecond)

	_, err := c.Get("k")
	assert.Equal(t, ErrCacheExited, err)
	_, err = c.Set("k", "v")
	assert.Equal(t, ErrCacheExited, err)
	err = c.Del("k")
	assert.Equal(t, ErrCacheExited, err)
	time.Sleep(shutdownWindow + 10*time.Millisecond)
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
