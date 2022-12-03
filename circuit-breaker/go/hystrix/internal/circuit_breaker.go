package internal

import (
	"hystrix/config"
	"hystrix/internal/collector"
	"hystrix/internal/command"
	"sync"
	"time"
)

var (
	circuitBreakersMutex sync.RWMutex
	circuitBreakers      map[string]*CircuitBreaker
)

func init() {
	circuitBreakers = make(map[string]*CircuitBreaker)
}

// GetCircuitBreaker tries to find a CircuitBreaker associated with given name and
// create one if not found.
//
// It's designed to be thread-safe.
func GetCircuitBreaker(name string) (*CircuitBreaker, bool, error) {
	circuitBreakersMutex.RLock()
	if cb, ok := circuitBreakers[name]; ok {
		circuitBreakersMutex.RUnlock()
		return cb, false, nil
	}
	circuitBreakersMutex.RUnlock()
	circuitBreakersMutex.Lock()
	defer circuitBreakersMutex.Unlock()
	// Check again in case the circuit breaker has been created by another thread.
	if cb, ok := circuitBreakers[name]; ok {
		return cb, false, nil
	}
	circuitBreakers[name] = newCircuitBreaker(name)
	return circuitBreakers[name], true, nil
}

type CircuitBreaker struct {
	sync.Mutex
	name               string
	open               bool
	metricBroker       MetricBroker
	lastTransitionTime time.Time
}

func newCircuitBreaker(name string) *CircuitBreaker {
	memoryCollector := collector.NewMemoryCollector(name)

	return &CircuitBreaker{
		name:         name,
		metricBroker: NewChannelBroker(memoryCollector),
	}
}

// Allow decides whether CircuitBreaker is closed or not.
func (cb *CircuitBreaker) Allow() bool {
	return !cb.isOpen() || cb.isHalfOpen()
}

func (cb *CircuitBreaker) isOpen() bool {
	cb.Lock()
	defer cb.Unlock()

	if cb.open {
		return true
	}

	snapshot := cb.metricBroker.Collector().Snapshot()
	// TODO: configurable least attempts to count
	if snapshot.Requests < config.DefaultMinRequestNum {
		return false
	}
	// TODO: configurable error percent
	percent := int(float64(snapshot.Errors) / float64(snapshot.Requests) * 100)
	if percent > config.DefaultErrorPercentThreshold {
		cb.lastTransitionTime = time.Now()
		cb.open = true
	}
	return cb.open
}

func (cb *CircuitBreaker) isHalfOpen() bool {
	cb.Lock()
	defer cb.Unlock()
	// TODO: configuration sleep time.
	if cb.open && (time.Since(cb.lastTransitionTime) > time.Duration(config.DefaultBackoffMillis)*time.Millisecond) {
		cb.lastTransitionTime = time.Now()
		return true
	}
	return false
}

// Report sends the execution metrics to collectors asynchronously.
func (cb *CircuitBreaker) Report(execution *command.Execution) error {
	cb.Lock()
	defer cb.Unlock()

	if cb.open && execution.Status == command.ExecutionStatusSuccess {
		cb.open = false
		cb.metricBroker.Reset()
	}

	return cb.metricBroker.Report(execution)
}
