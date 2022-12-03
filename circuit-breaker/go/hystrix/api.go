package hystrix

import (
	"context"
	"errors"
	"hystrix/config"
	"hystrix/internal"
	"hystrix/internal/command"
	"log"
	"sync"
	"time"
)

type runFunc func(context.Context) error
type fallbackFunc func(context.Context, error) error

var noopFallback = func(ctx context.Context, runErr error) error { return runErr }

var (
	// ErrCircuitBreakerOpen returns when an execution attempt "short circuits". This happens due to the circuit being measured as unhealthy.
	ErrCircuitBreakerOpen = errors.New("circuit open")
	// ErrTimeout occurs when the provided function takes too long to execute.
	ErrTimeout = errors.New("timeout")
)

func Go(ctx context.Context, name string, run runFunc, fallback fallbackFunc) (errChan chan error) {
	errChan = make(chan error, 1)
	execution := command.NewExecution()

	circuitBreaker, _, err := internal.GetCircuitBreaker(name)
	if err != nil {
		errChan <- err
		return
	}

	final := new(sync.Once)
	// fallbackWithError should only be called inside final.Do
	fallbackWithError := func(execErr error) {
		if fallback == nil {
			fallback = noopFallback
		}
		fbErr := fallback(ctx, execErr)
		if fbErr != nil {
			execution.FallbackStatus = command.ExecutionStatusFailure
		} else {
			execution.FallbackStatus = command.ExecutionStatusSuccess
		}
		errChan <- fbErr
	}
	report := func(execution *command.Execution) {
		if err = circuitBreaker.Report(execution); err != nil {
			log.Printf("report err %v\n", err)
		}
	}

	finChan := make(chan interface{}, 1)
	go func() {
		defer func() { finChan <- struct{}{} }()
		if !circuitBreaker.Allow() {
			final.Do(func() {
				execution.Status = command.ExecutionStatusShortCircuit
				fallbackWithError(ErrCircuitBreakerOpen)
				report(execution)
			})
			return
		}

		execution.Start()
		runErr := run(ctx)
		execution.Finish()

		final.Do(func() {
			if runErr != nil {
				execution.Status = command.ExecutionStatusFailure
				fallbackWithError(runErr)
			} else {
				execution.Status = command.ExecutionStatusSuccess
				errChan <- runErr
			}
			report(execution)
		})
	}()

	go func() {
		// TODO: configurable window size.
		timer := time.NewTimer(time.Duration(config.DefaultTimeoutMillis) * time.Millisecond)
		defer timer.Stop()

		select {
		case <-finChan:
			// final has been executed in another goroutine
		case <-ctx.Done():
			final.Do(func() {
				execution.Status = command.ExecutionStatusFailure
				fallbackWithError(ctx.Err())
			})
		case <-timer.C:
			final.Do(func() {
				execution.Status = command.ExecutionStatusTimeout
				fallbackWithError(ErrTimeout)
			})
		}
		return
	}()

	return errChan
}

func Do(ctx context.Context, name string, run runFunc, fallback fallbackFunc) error {
	done := make(chan struct{}, 1)
	runF := func(ctx context.Context) error {
		if err := run(ctx); err != nil {
			return err
		}
		done <- struct{}{}
		return nil
	}

	if fallback == nil {
		fallback = noopFallback
	}

	fallbackF := func(ctx context.Context, runErr error) error {
		if err := fallback(ctx, runErr); err != nil {
			return err
		}
		done <- struct{}{}
		return nil
	}

	errChan := Go(ctx, name, runF, fallbackF)
	select {
	case <-done:
		return nil
	case err := <-errChan:
		return err
	}
}
