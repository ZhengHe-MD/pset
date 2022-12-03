package command

import "time"

type ExecutionStatus int

const (
	ExecutionStatusUnspecified ExecutionStatus = iota
	ExecutionStatusSuccess
	ExecutionStatusFailure
	ExecutionStatusShortCircuit
	ExecutionStatusTimeout
)

// Execution keeps information about an execution of a command.
type Execution struct {
	start          time.Time
	Status         ExecutionStatus
	FallbackStatus ExecutionStatus
	Duration       time.Duration
}

// NewExecution generates a new Execution instance.
func NewExecution() *Execution {
	return &Execution{
		Status:         ExecutionStatusUnspecified,
		FallbackStatus: ExecutionStatusUnspecified,
	}
}

// Start denotes that the execution has started.
func (e *Execution) Start() {
	e.start = time.Now()
}

// Finish denotes that the execution has finished.
func (e *Execution) Finish() {
	e.Duration = time.Since(e.start)
}
