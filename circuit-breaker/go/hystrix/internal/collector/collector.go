package collector

import (
	"time"
)

// Snapshot represents a snapshot of metrics in a Collector within a valid window.
type Snapshot struct {
	Requests      int // number of requests
	Errors        int // number of errors, including errors before and after execution.
	Successes     int // number of successes
	Failures      int // number of failures, only errors during execution counts.
	ShortCircuits int // number of times that the execution has been short-circuited.
}

// Sample represents the data of one execution.
type Sample struct {
	Requests      int
	Errors        int
	Successes     int
	Failures      int
	ShortCircuits int
	Duration      time.Duration
}

// Interface represents the contract conformed by all concrete collectors.
type Interface interface {
	// Collect takes an Execution record and merge it with history.
	Collect(Sample)
	// Reset clears all history.
	Reset()
	// Snapshot takes a snapshot of current metrics.
	Snapshot() Snapshot
}
