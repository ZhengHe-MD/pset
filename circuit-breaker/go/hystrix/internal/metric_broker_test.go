package internal

import (
	"hystrix/internal/collector"
	"hystrix/internal/command"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChannelBroker_All(t *testing.T) {
	mc := collector.NewMemoryCollector("")
	cb := NewChannelBroker(mc)

	{
		assert.NoError(t, cb.Report(&command.Execution{
			Status: command.ExecutionStatusSuccess,
		}))
		time.Sleep(5 * time.Millisecond)
		ss := cb.Collector().Snapshot()
		assert.Equal(t, 1, ss.Successes)
		assert.Equal(t, 1, ss.Requests)
		assert.Equal(t, 0, ss.Errors)
		assert.Equal(t, 0, ss.Failures)
	}

	{
		assert.NoError(t, cb.Report(&command.Execution{
			Status: command.ExecutionStatusTimeout,
		}))
		time.Sleep(5 * time.Millisecond)
		ss := cb.Collector().Snapshot()
		assert.Equal(t, 1, ss.Successes)
		assert.Equal(t, 2, ss.Requests)
		assert.Equal(t, 1, ss.Errors)
		assert.Equal(t, 1, ss.Failures)
	}

	{
		assert.NoError(t, cb.Report(&command.Execution{
			Status: command.ExecutionStatusShortCircuit,
		}))
		time.Sleep(5 * time.Millisecond)
		ss := cb.Collector().Snapshot()
		assert.Equal(t, 1, ss.Successes)
		assert.Equal(t, 3, ss.Requests)
		assert.Equal(t, 2, ss.Errors)
		assert.Equal(t, 1, ss.Failures)
	}

	{
		assert.NoError(t, cb.Report(&command.Execution{
			Status: command.ExecutionStatusFailure,
		}))
		time.Sleep(5 * time.Millisecond)
		ss := cb.Collector().Snapshot()
		assert.Equal(t, 1, ss.Successes)
		assert.Equal(t, 4, ss.Requests)
		assert.Equal(t, 3, ss.Errors)
		assert.Equal(t, 2, ss.Failures)
	}

	close(cb.executionCh)
}
