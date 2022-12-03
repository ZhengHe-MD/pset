package collector

import (
	"hystrix/internal/window"
)

var _ Interface = (*MemoryCollector)(nil)

// MemoryCollector is an implementation of Interface by simply keeping
// metric data in memory.
type MemoryCollector struct {
	requests      *window.Counter
	successes     *window.Counter
	failures      *window.Counter
	errors        *window.Counter
	shortCircuits *window.Counter
}

// NewMemoryCollector is the Initializer of MemoryCollector.
func NewMemoryCollector(name string) *MemoryCollector {
	mc := &MemoryCollector{}
	mc.Reset()
	return mc
}

func (m *MemoryCollector) Collect(metrics Sample) {
	m.requests.Inc(metrics.Requests)
	m.successes.Inc(metrics.Successes)
	m.failures.Inc(metrics.Failures)
	m.errors.Inc(metrics.Errors)
	m.shortCircuits.Inc(metrics.ShortCircuits)
}

func (m *MemoryCollector) Reset() {
	m.requests = window.NewCounter()
	m.successes = window.NewCounter()
	m.failures = window.NewCounter()
	m.errors = window.NewCounter()
	m.shortCircuits = window.NewCounter()
}

func (m *MemoryCollector) Snapshot() Snapshot {
	return Snapshot{
		Requests:      m.requests.Sum(),
		Errors:        m.errors.Sum(),
		Successes:     m.successes.Sum(),
		Failures:      m.failures.Sum(),
		ShortCircuits: m.shortCircuits.Sum(),
	}
}
