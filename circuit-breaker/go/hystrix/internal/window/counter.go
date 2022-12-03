package window

import (
	"sync"
	"time"
)

// Counter tracks the number of events in a time window,
//
//	which is a 10-second fixed time window for now.
type Counter struct {
	// TODO: test performance boost on switching to RWMutex
	sync.Mutex
	buckets map[int64]int
}

// NewCounter initializes a Counter.
func NewCounter() *Counter {
	return &Counter{
		buckets: make(map[int64]int),
	}
}

// Inc add a positive number to current bucket.
func (c *Counter) Inc(n int) {
	if n <= 0 {
		return
	}

	bucket := time.Now().Unix()
	c.Lock()
	defer c.Unlock()
	c.buckets[bucket] += n
	c.removeOutdatedBuckets()
}

// Sum sums the counts over buckets in bound.
func (c *Counter) Sum() (n int) {
	c.Lock()
	defer c.Unlock()

	lb := time.Now().Unix() - 10
	for ts, cnt := range c.buckets {
		if ts > lb {
			n += cnt
		}
	}
	return
}

// should be called inside critical area.
func (c *Counter) removeOutdatedBuckets() {
	lb := time.Now().Unix() - 10

	for ts := range c.buckets {
		if ts <= lb {
			delete(c.buckets, ts)
		}
	}
}
