package internal

import (
	"errors"
	"hystrix/internal/collector"
	"hystrix/internal/command"
	"log"
)

// MetricBroker separates the concerns of metric collection and command execution.
// It makes metric collection asynchronous.
type MetricBroker interface {
	// Report sends a command's execution info asynchronously.
	Report(*command.Execution) error
	// Reset clears history.
	Reset()
	// Collector returns the underlying collector.
	Collector() collector.Interface
}

var _ MetricBroker = (*ChannelBroker)(nil)

type ChannelBroker struct {
	executionCh chan *command.Execution
	collector   collector.Interface
}

func NewChannelBroker(collector collector.Interface) *ChannelBroker {
	cb := &ChannelBroker{
		executionCh: make(chan *command.Execution, 2000), // TODO: use a larger number if necessary
		collector:   collector,
	}
	cb.Reset()
	go cb.monitor()
	return cb
}

func (c *ChannelBroker) monitor() {
	for execution := range c.executionCh {
		sample := collector.Sample{Requests: 1}
		switch execution.Status {
		case command.ExecutionStatusSuccess:
			sample.Successes = 1
		case command.ExecutionStatusFailure:
			sample.Failures = 1
			sample.Errors = 1
		case command.ExecutionStatusTimeout:
			sample.Failures = 1
			sample.Errors = 1
		case command.ExecutionStatusShortCircuit:
			sample.ShortCircuits = 1
			sample.Errors = 1
		default:
			log.Printf("invalid execution, not reachable, %#v\n", sample)
			continue
		}
		c.Collector().Collect(sample)
	}
}

func (c *ChannelBroker) Report(execution *command.Execution) error {
	select {
	case c.executionCh <- execution:
		return nil
	default:
		return errors.New("metrics channel is at capacity")
	}
}

func (c *ChannelBroker) Reset() {
	c.collector.Reset()
}

func (c *ChannelBroker) Collector() collector.Interface {
	return c.collector
}
