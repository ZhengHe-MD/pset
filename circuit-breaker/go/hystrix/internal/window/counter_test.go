package window

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCounter_Inc(t *testing.T) {
	c := NewCounter()
	bkt := time.Now().Unix()
	c.Inc(1)
	assert.Equal(t, 1, c.buckets[bkt])
	time.Sleep(time.Second)
	c.Inc(2)
	assert.Equal(t, 1, c.buckets[bkt])
	assert.Equal(t, 2, c.buckets[bkt+1])
	time.Sleep(9 * time.Second)
	c.Inc(1)
	assert.Equal(t, 0, c.buckets[bkt])
	assert.Equal(t, 1, c.buckets[bkt+10])
}

func TestCounter_Sum(t *testing.T) {
	bkt := time.Now().Unix()
	c := NewCounter()
	c.buckets = map[int64]int{
		bkt - 10: 3,
		bkt - 9:  5,
		bkt - 4:  2,
	}

	assert.Equal(t, 7, c.Sum())
	time.Sleep(time.Second)
	assert.Equal(t, 2, c.Sum())
}

func BenchmarkCounter_Inc(b *testing.B) {
	c := NewCounter()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Inc(1)
	}
}
