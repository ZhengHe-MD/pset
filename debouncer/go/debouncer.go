package debouncer

import (
	"sync"
	"time"
)

type Interface interface {
	Debounce(f func())
}

func New(after time.Duration) Interface {
	return &debouncer{after: after}
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

func (d *debouncer) Debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}
