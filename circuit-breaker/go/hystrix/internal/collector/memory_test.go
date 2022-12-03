package collector

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewMemoryCollector(t *testing.T) {
	mc := NewMemoryCollector("")
	assert.NotNil(t, mc.requests)
	assert.Equal(t, 0, mc.requests.Sum())
	assert.NotNil(t, mc.successes)
	assert.Equal(t, 0, mc.successes.Sum())
	assert.NotNil(t, mc.failures)
	assert.Equal(t, 0, mc.failures.Sum())
	assert.NotNil(t, mc.errors)
	assert.Equal(t, 0, mc.errors.Sum())
	assert.NotNil(t, mc.shortCircuits)
	assert.Equal(t, 0, mc.shortCircuits.Sum())
}

func TestMemoryCollector_Collect(t *testing.T) {
	mc := NewMemoryCollector("")
	sp := Sample{
		Requests:      12,
		Errors:        9,
		Successes:     3,
		Failures:      4,
		ShortCircuits: 5,
		Duration:      time.Second,
	}
	mc.Collect(sp)
	assert.Equal(t, sp.Requests, mc.requests.Sum())
	assert.Equal(t, sp.Errors, mc.errors.Sum())
	assert.Equal(t, sp.Successes, mc.successes.Sum())
	assert.Equal(t, sp.Failures, mc.failures.Sum())
	assert.Equal(t, sp.ShortCircuits, mc.shortCircuits.Sum())
}

func TestMemoryCollector_Snapshot(t *testing.T) {
	mc := NewMemoryCollector("")
	sp := Sample{
		Requests:      12,
		Errors:        9,
		Successes:     3,
		Failures:      4,
		ShortCircuits: 5,
		Duration:      time.Second,
	}
	mc.Collect(sp)
	ss := mc.Snapshot()
	assert.Equal(t, sp.Requests, ss.Requests)
	assert.Equal(t, sp.Errors, ss.Errors)
	assert.Equal(t, sp.Successes, ss.Successes)
	assert.Equal(t, sp.Failures, ss.Failures)
	assert.Equal(t, sp.ShortCircuits, ss.ShortCircuits)
}
