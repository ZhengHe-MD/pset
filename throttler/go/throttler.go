package throttler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Interface interface {
	// Throttle is a thread-safe method supposed to be called by
	// multiple goroutines. Every caller provides a resource key
	// 'k' and a user-defined function `f`. Caller should deal
	// with two cases, being the first or the rest. The first
	// caller should go and get the data and others should just
	// fetch the data locally. It's guaranteed that other callers'
	// `f` will only be called after the first one finishes its work.
	//
	// Example:
	// k := "uid-13795"
	// t := New(5 * time.Second)
	// t.Throttle(ctx, k, func(first bool) error {
	//   if !first {
	//     return cache[k], nil
	//   }
	//   d := getData()
	//   cache[k] = d
	//   return d, nil
	// })
	Throttle(ctx context.Context, k string, f func(first bool) error) error
}

func New(ctx context.Context, timeout time.Duration) Interface {
	t := &throttler{
		requestCh: make(chan *request),
		resources: make(map[string]*lock),
		timeout:   timeout,
	}

	go t.serve(ctx)
	go t.gc(ctx, 1*time.Minute)

	return t
}

type throttler struct {
	requestCh chan *request
	resources map[string]*lock
	timeout   time.Duration
	exited    atomic.Bool
}

type lock struct {
	expiry time.Time
	cond   *sync.Cond
}

func (l *lock) hasExpired() bool {
	return l.expiry.Before(time.Now())
}

type requestType int

const (
	unspecified requestType = iota
	get
	del
	exp
	exit
)

type request struct {
	typ    requestType
	key    string
	respCh chan *response
}

type response struct {
	first bool
	cond  *sync.Cond
}

func (t *throttler) Throttle(ctx context.Context, key string, f func(first bool) error) (err error) {
	if t.exited.Load() {
		return errors.New("throttle on an exited throttler")
	}

	r := &request{
		typ:    get,
		key:    key,
		respCh: make(chan *response),
	}
	t.requestCh <- r

	resp := <-r.respCh

	defer func() {
		if rc := recover(); rc != nil {
			err = fmt.Errorf("%v", rc)
		}
	}()

	resp.cond.L.Lock()
	if !resp.first {
		resp.cond.Wait()
		err = f(resp.first)
		resp.cond.L.Unlock()
		return
	}
	resp.cond.L.Unlock()

	defer func() {
		go func() {
			t.requestCh <- &request{typ: del, key: key}
		}()
	}()

	return f(resp.first)
}

func (t *throttler) serve(ctx context.Context) {
	for {
		r := <-t.requestCh

		switch r.typ {
		case get:
			t.get(r)
		case del:
			t.del(r.key)
		case exp:
			t.exp()
		case exit:
			t.exited.Store(true)
			return
		}
	}
}

func (t *throttler) get(r *request) {
	l, ok := t.resources[r.key]
	if !ok {
		l = &lock{
			expiry: time.Now().Add(t.timeout),
			cond:   sync.NewCond(&sync.Mutex{}),
		}
		t.resources[r.key] = l
		r.respCh <- &response{first: true, cond: l.cond}
		return
	}

	if l.hasExpired() {
		t.del(r.key)
		t.get(r)
		return
	}

	r.respCh <- &response{first: false, cond: l.cond}
}

func (t *throttler) del(key string) {
	l, ok := t.resources[key]
	if !ok {
		return
	}
	l.cond.Broadcast()
	delete(t.resources, key)
}

func (t *throttler) exp() {
	for k, l := range t.resources {
		if l.hasExpired() {
			t.del(k)
		}
	}
}

var (
	expireRequest = &request{typ: exp}
	exitRequest   = &request{typ: exit}
)

func (t *throttler) gc(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			t.requestCh <- exitRequest
			return
		case <-ticker.C:
			t.requestCh <- expireRequest
		}
	}
}
