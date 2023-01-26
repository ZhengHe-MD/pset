package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

type Interface interface {
	// Get returns the value corresponding to the given key
	Get(key interface{}) (val interface{}, err error)
	// Set stores key-val pair in the cache and returns the previous value
	// associated with the key if it exists.
	Set(key, val interface{}) (preVal interface{}, err error)
	// Del removes key and its value from the cache
	Del(key interface{}) (err error)
}

type (
	cache struct {
		lookup       map[interface{}]*value
		updateExpiry time.Duration
		requestCh    chan *request
		exited       atomic.Bool
		gcInterval   time.Duration
	}

	value struct {
		updateTime time.Time
		data       interface{}
	}

	request struct {
		typ        requestType
		key        interface{}
		val        interface{}
		responseCh chan *response
	}

	response struct {
		err error
		val interface{}
	}

	requestType int
)

var (
	ErrCacheExited = errors.New("cache service exited")
	ErrNotFound    = errors.New("key not found")
)

const (
	unspecified requestType = iota
	get
	set
	del
	exp
	exit
)

func New(ctx context.Context, updateExpiry time.Duration) Interface {
	c := &cache{
		lookup:       make(map[interface{}]*value),
		updateExpiry: updateExpiry,
		requestCh:    make(chan *request),
		gcInterval:   1 * time.Minute,
	}

	go c.serve()
	go c.exit(ctx)
	go c.gc(ctx)

	return c
}

func (c *cache) Get(key interface{}) (val interface{}, err error) {
	if c.exited.Load() {
		return nil, ErrCacheExited
	}

	responseCh := make(chan *response)

	c.requestCh <- &request{
		typ:        get,
		key:        key,
		responseCh: responseCh,
	}

	resp := <-responseCh
	return resp.val, resp.err
}

func (c *cache) Set(key, val interface{}) (currVal interface{}, err error) {
	if c.exited.Load() {
		return nil, ErrCacheExited
	}

	responseCh := make(chan *response)
	c.requestCh <- &request{
		typ:        set,
		key:        key,
		val:        val,
		responseCh: responseCh,
	}
	resp := <-responseCh
	return resp.val, resp.err
}

func (c *cache) Del(key interface{}) (err error) {
	if c.exited.Load() {
		return ErrCacheExited
	}

	responseCh := make(chan *response)
	c.requestCh <- &request{
		typ:        del,
		key:        key,
		responseCh: responseCh,
	}
	resp := <-responseCh
	return resp.err
}

func (c *cache) serve() {
	for {
		r := <-c.requestCh
		resp := &response{}

		switch r.typ {
		case get:
			val, ok := c.lookup[r.key]
			if !ok || val.expired(c.updateExpiry) {
				resp.err = ErrNotFound
			} else {
				resp.val = val.data
			}
			r.responseCh <- resp
		case set:
			preVal, ok := c.lookup[r.key]
			c.lookup[r.key] = &value{
				updateTime: time.Now(),
				data:       r.val,
			}
			if ok {
				resp.val = preVal.data
			}
			r.responseCh <- resp
		case del:
			_, ok := c.lookup[r.key]
			if !ok {
				resp.err = ErrNotFound
			} else {
				delete(c.lookup, r.key)
			}
			r.responseCh <- resp
		case exp:
			for k, v := range c.lookup {
				if v.expired(c.updateExpiry) {
					delete(c.lookup, k)
				}
			}
			r.responseCh <- resp
		case exit:
			return
		}
	}
}

const shutdownWindow = 1 * time.Second

func (c *cache) exit(ctx context.Context) {
	<-ctx.Done()
	c.exited.Store(true)
	time.Sleep(shutdownWindow)
	c.requestCh <- &request{typ: exit}
}

func (c *cache) gc(ctx context.Context) {
	// no expiration, no garbage collection
	if c.updateExpiry <= 0 {
		return
	}

	ticker := time.NewTicker(c.gcInterval)
	defer ticker.Stop()

	for {
		if c.exited.Load() {
			break
		}

		select {
		case <-ctx.Done():
			break
		case <-ticker.C:
			responseCh := make(chan *response)
			c.requestCh <- &request{typ: exp, responseCh: responseCh}
			<-responseCh
		}
	}
}

func (v *value) expired(timeout time.Duration) bool {
	return timeout > 0 && time.Since(v.updateTime) > timeout
}
