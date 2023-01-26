package internal

import (
	"hystrix/config"
	"hystrix/internal/collector"
	"hystrix/internal/command"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCircuitBreaker(t *testing.T) {
	cb1, created1, err1 := GetCircuitBreaker("")
	assert.NotNil(t, cb1)
	assert.True(t, created1)
	assert.NoError(t, err1)

	cb2, created2, err2 := GetCircuitBreaker("")
	assert.NotNil(t, cb2)
	assert.False(t, created2)
	assert.NoError(t, err2)

	assert.Same(t, cb1, cb2)
}

func TestCircuitBreaker_Allow(t *testing.T) {
	t.Run("should allow the first call", func(t *testing.T) {
		cb, _, _ := GetCircuitBreaker("1")
		assert.True(t, cb.Allow())
	})

	t.Run("should allow the first few calls no matter what happened", func(t *testing.T) {
		cb, _, _ := GetCircuitBreaker("2")
		execution := &command.Execution{Status: command.ExecutionStatusFailure}
		for i := 0; i < config.DefaultMinRequestNum-1; i++ {
			assert.NoError(t, cb.Report(execution))
		}
		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, config.DefaultMinRequestNum-1, cb.metricBroker.Collector().Snapshot().Failures)
		assert.True(t, cb.Allow())
	})

	t.Run("should open after too many continuous failures", func(t *testing.T) {
		cb, _, _ := GetCircuitBreaker("3")
		execution := &command.Execution{Status: command.ExecutionStatusFailure}
		for i := 0; i < config.DefaultMinRequestNum+1; i++ {
			assert.NoError(t, cb.Report(execution))
		}
		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, config.DefaultMinRequestNum+1, cb.metricBroker.Collector().Snapshot().Failures)
		assert.False(t, cb.Allow())
	})

	t.Run("should allow after it stays open longer than the sleep time window", func(t *testing.T) {
		cb, _, _ := GetCircuitBreaker("4")
		execution := &command.Execution{Status: command.ExecutionStatusFailure}
		for i := 0; i < 21; i++ {
			assert.NoError(t, cb.Report(execution))
		}
		time.Sleep(5 * time.Millisecond)
		assert.False(t, cb.Allow())
		time.Sleep(5 * time.Second)
		assert.True(t, cb.Allow())

		cb.Report(&command.Execution{Status: command.ExecutionStatusSuccess})
		ss := cb.metricBroker.Collector().Snapshot()
		assert.Equal(t, collector.Snapshot{}, ss)
	})
}
