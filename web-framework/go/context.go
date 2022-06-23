package main

import (
	"context"
	"net/http"
	"time"
)

var _ context.Context = (*Context)(nil)

// Context allows us to share pre-request variables among middlewares,
// provide easy ways to access query and path parameters.
type Context struct {
	ctx    context.Context
	w      http.ResponseWriter
	r      *http.Request
	params map[string]string
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		ctx: context.Background(),
		w:   w,
		r:   r,
	}
}

func (c *Context) Query(key string) string {
	return c.r.URL.Query().Get(key)
}

func (c *Context) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Context) Err() error {
	return c.ctx.Err()
}

func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}
