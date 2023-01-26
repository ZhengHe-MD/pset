package debouncer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebouncer_Debounce(t *testing.T) {
	t.Run("single function", func(t *testing.T) {
		var cnt uint64

		f1 := func() {
			atomic.AddUint64(&cnt, 1)
		}

		d := New(100 * time.Millisecond)

		for i := 0; i < 3; i++ {
			for j := 0; j < 10; j++ {
				d.Debounce(f1)
			}
			time.Sleep(200 * time.Millisecond)
		}

		assert.Equal(t, 3, int(atomic.LoadUint64(&cnt)))
	})

	t.Run("last call wins", func(t *testing.T) {
		var cnt uint64

		f1 := func() { atomic.AddUint64(&cnt, 1) }
		f2 := func() { atomic.AddUint64(&cnt, 2) }

		d := New(100 * time.Millisecond)

		for i := 0; i < 4; i++ {
			for j := 0; j < 10; j++ {
				d.Debounce(f1)
			}
			for j := 0; j < 10; j++ {
				d.Debounce(f2)
			}
			time.Sleep(200 * time.Millisecond)
		}

		assert.Equal(t, 8, int(atomic.LoadUint64(&cnt)))
	})

	t.Run("non-blocking concurrent calls", func(t *testing.T) {
		var wg sync.WaitGroup
		var flag uint64

		d := New(100 * time.Millisecond)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				d.Debounce(func() {
					atomic.CompareAndSwapUint64(&flag, 0, 1)
				})
			}()
		}

		wg.Wait()

		time.Sleep(500 * time.Millisecond)
		assert.Equal(t, 1, int(atomic.LoadUint64(&flag)))
	})
}

func BenchmarkDebounce(b *testing.B) {
	var cnt uint64

	f := func() {
		atomic.AddUint64(&cnt, 1)
	}

	d := New(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Debounce(f)
	}

	assert.Equal(b, 0, int(atomic.LoadUint64(&cnt)))
}
