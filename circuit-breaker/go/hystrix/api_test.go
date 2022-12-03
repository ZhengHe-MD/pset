package hystrix

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"hystrix/config"
	"testing"
	"time"
)

func TestSuccess(t *testing.T) {
	resultCh := make(chan int, 1)
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		resultCh <- 1
		return nil
	}, nil)

	assert.NoError(t, err)
	assert.Equal(t, 1, <-resultCh)
}

func TestFallback(t *testing.T) {
	resultCh := make(chan int, 1)
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		return errors.New("error")
	}, func(ctx context.Context, err error) error {
		if err.Error() == "error" {
			resultCh <- 1
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, <-resultCh)
}

func TestTimeout(t *testing.T) {
	resultChan := make(chan int, 1)
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		resultChan <- 1
		return nil
	}, func(ctx context.Context, err error) error {
		if err == ErrTimeout {
			resultChan <- 2
		}
		return nil
	})

	assert.Equal(t, 2, <-resultChan)
	assert.NoError(t, err)
}

func TestTimeoutEmptyFallback(t *testing.T) {
	resultChan := make(chan int, 1)
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		resultChan <- 1
		return nil
	}, nil)

	assert.Equal(t, ErrTimeout, err)
}

func TestNilFallbackRunError(t *testing.T) {
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		return errors.New("run_error")
	}, nil)
	assert.Equal(t, "run_error", err.Error())
}

func TestFailedFallback(t *testing.T) {
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		return errors.New("run_error")
	}, func(ctx context.Context, err error) error {
		return errors.New("fallback_error")
	})
	assert.Equal(t, "fallback_error", err.Error())
}

func TestCloseCircuitAfterSuccess(t *testing.T) {
	for i := 0; i < config.DefaultMinRequestNum; i++ {
		err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
			return errors.New("run_error")
		}, nil)
		assert.Equal(t, "run_error", err.Error())
	}

	time.Sleep(5 * time.Millisecond)
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		return errors.New("run_error")
	}, nil)
	assert.Equal(t, ErrCircuitBreakerOpen, err)

	time.Sleep(time.Duration(config.DefaultBackoffMillis) * time.Millisecond)
	err = Do(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	}, nil)
	assert.NoError(t, err)
}

func TestFailAfterTimeout(t *testing.T) {
	fallbackResultCh := make(chan int, 2)

	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		return errors.New("run_error")
	}, func(ctx context.Context, err error) error {
		fallbackResultCh <- 1
		return err
	})
	assert.Equal(t, ErrTimeout, err)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 1, len(fallbackResultCh))
}

func TestSlowFallbackOpenCircuit(t *testing.T) {
	// open circuit
	for i := 0; i < config.DefaultMinRequestNum; i++ {
		err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
			return errors.New("run_error")
		}, nil)
		assert.Equal(t, "run_error", err.Error())
	}

	time.Sleep(5 * time.Millisecond)
	// slow fallback
	fallbackResultCh := make(chan int, 2)
	err := Do(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	}, func(ctx context.Context, err error) error {
		time.Sleep(2 * time.Second)
		fallbackResultCh <- 1
		return nil
	})

	assert.NoError(t, err)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 1, len(fallbackResultCh))
}

func TestContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	errCh := Go(ctx, t.Name(), func(ctx context.Context) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}, nil)

	cancel()
	assert.Equal(t, context.Canceled, <-errCh)
}

func TestContextDeadlineExceeded(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
	errCh := Go(ctx, t.Name(), func(ctx context.Context) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}, nil)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, context.DeadlineExceeded, <-errCh)
}
