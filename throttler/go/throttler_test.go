package throttler

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func _newWithGCInterval(ctx context.Context, timeout time.Duration, interval time.Duration) *throttler {
	t := &throttler{
		requestCh: make(chan *request),
		resources: make(map[string]*lock),
		timeout:   timeout,
	}

	go t.serve(ctx)
	go t.gc(ctx, interval)

	return t
}

func gracefullyCancel(cancel context.CancelFunc) {
	// let goroutines exit
	time.Sleep(100 * time.Millisecond)
	cancel()
}

func TestThrottler_Throttle(t *testing.T) {
	t.Run("only one goroutine passes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer gracefullyCancel(cancel)
		tt := New(ctx, 10*time.Second)
		wg := sync.WaitGroup{}
		var numPass uint64
		var numBlock uint64
		cache := make(map[string]int)
		key, val := "uid", 1234
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				_ = tt.Throttle(context.Background(), key, func(first bool) error {
					if first {
						atomic.AddUint64(&numPass, 1)
						time.Sleep(100 * time.Millisecond)
						cache[key] = val
					} else {
						atomic.AddUint64(&numBlock, 1)
					}

					assert.Equal(t, val, cache[key])
					return nil
				})
				wg.Done()
			}()
		}
		wg.Wait()
		assert.Equal(t, 1, int(numPass))
		assert.Equal(t, 9, int(numBlock))
	})

	t.Run("the first goroutine timeouts", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer gracefullyCancel(cancel)
		tt := New(ctx, 100*time.Millisecond)
		wg := sync.WaitGroup{}
		cache := make(map[string]int)
		key := "uid"
		var timeoutErrNum, notFoundErrNum uint64
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				err := tt.Throttle(context.Background(), key, func(first bool) error {
					if first {
						time.Sleep(200 * time.Millisecond)
						return errors.New("timeout")
					}

					if _, ok := cache[key]; !ok {
						return errors.New("not found")
					}
					return nil
				})
				if err.Error() == "timeout" {
					atomic.AddUint64(&timeoutErrNum, 1)
				} else {
					atomic.AddUint64(&notFoundErrNum, 1)
				}
				wg.Done()
			}()
		}
		go func() {
			time.Sleep(150 * time.Millisecond)
			_ = tt.Throttle(context.Background(), key, func(first bool) error {
				assert.True(t, first)
				return nil
			})
		}()
		wg.Wait()
		assert.Equal(t, 1, int(timeoutErrNum))
		assert.Equal(t, 9, int(notFoundErrNum))
		assert.NotContains(t, cache, key)
	})

	t.Run("gc", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer gracefullyCancel(cancel)
		tt := _newWithGCInterval(ctx, 100*time.Millisecond, 30*time.Millisecond)
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				_ = tt.Throttle(context.Background(), "k", func(first bool) error {
					if first {
						time.Sleep(200 * time.Millisecond)
						return errors.New("timeout")
					}
					return nil
				})
				wg.Done()
			}()
		}
		go func() {
			time.Sleep(150 * time.Millisecond)
			assert.Equal(t, len(tt.resources), 0)
		}()
		wg.Wait()
		assert.Equal(t, len(tt.resources), 0)
	})

	t.Run("throttle on an exited throttler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		tt := New(ctx, 100*time.Millisecond)
		cancel()
		// let the exit command execute
		time.Sleep(50 * time.Millisecond)
		// after exited, throttle is prohibited
		err := tt.Throttle(context.Background(), "k2", func(first bool) error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exited")
	})

	t.Run("panic f", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer gracefullyCancel(cancel)
		tt := New(ctx, 100*time.Millisecond)

		err := tt.Throttle(context.Background(), "k", func(first bool) error {
			panic("intentional panic")
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "intentional panic")
	})
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
