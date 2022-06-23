package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

type Handler func(*Context)

var _ http.Handler = (*Router)(nil)

var (
	defaultNotFoundHandle Handler = func(c *Context) {
		http.NotFoundHandler().ServeHTTP(c.w, c.r)
	}
	defaultMethodNotAllowedHandle Handler = func(c *Context) {
		http.Error(c.w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
	}
	defaultPanicHandle Handler = func(c *Context) {
		http.Error(c.w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}
)

// Router is a minimum implementation of a web framework,
// which is fully compatible to go http server.
//
// features:
// 	1. Shortcuts to build REST apis;
// 	2. Easy to access parameters from url and query string;
//  3. Default handlers for panics and errors.
type Router struct {
	trees map[string]*handleTree

	NotFoundHandle         Handler
	MethodNotAllowedHandle Handler
	PanicHandle            Handler
}

func NewRouter() *Router {
	return &Router{
		trees:                  make(map[string]*handleTree),
		NotFoundHandle:         defaultNotFoundHandle,
		MethodNotAllowedHandle: defaultMethodNotAllowedHandle,
		PanicHandle:            defaultPanicHandle,
	}
}

func (mr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)

	defer func() {
		if rcv := recover(); rcv != nil {
			debug.PrintStack()
			mr.PanicHandle(c)
		}
	}()

	method, path := r.Method, r.URL.Path
	tree, ok := mr.trees[method]
	if !ok {
		mr.NotFoundHandle(c)
		return
	}

	handle, params, found := tree.getHandle(path)
	if !found {
		var foundInAnother bool
		for name, anotherTree := range mr.trees {
			if name == method {
				continue
			}
			if _, _, foundInAnother = anotherTree.getHandle(path); foundInAnother {
				mr.MethodNotAllowedHandle(c)
			}
		}

		if !foundInAnother {
			mr.NotFoundHandle(c)
		}
		return
	}

	c.params = params
	handle(c)
}

func (mr *Router) handle(method, path string, handle Handler) {
	if handle == nil {
		panic("handle should not be nil")
	}

	tree, ok := mr.trees[method]
	if !ok {
		tree = newHandleTree(method)
		mr.trees[method] = tree
	}

	tree.addHandle(path, handle)
}

func (mr *Router) GET(path string, handle Handler) {
	mr.handle(http.MethodGet, path, handle)
}

func (mr *Router) POST(path string, handle Handler) {
	mr.handle(http.MethodPost, path, handle)
}

func (mr *Router) PATCH(path string, handle Handler) {
	mr.handle(http.MethodPatch, path, handle)
}

func (mr *Router) PUT(path string, handle Handler) {
	mr.handle(http.MethodPut, path, handle)
}

func (mr *Router) DELETE(path string, handle Handler) {
	mr.handle(http.MethodDelete, path, handle)
}

func Query(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func Param(r *http.Request, key string) string {
	panic("implement me")
}

// Run actually calls ListenAndServe inside
func (mr *Router) Run(address string) error {
	fmt.Println("Start web-framework")

	for _, tree := range mr.trees {
		tree.desc()
	}

	return http.ListenAndServe(address, mr)
}

type handleTree struct {
	name     string
	mappings map[string]Handler
}

func newHandleTree(name string) *handleTree {
	return &handleTree{
		name:     name,
		mappings: make(map[string]Handler),
	}
}

func (ht *handleTree) addHandle(path string, handle Handler) {
	if _, registered := ht.mappings[path]; registered {
		panic(fmt.Sprintf("path %s registered in %s tree", path, ht.name))
	}

	ht.mappings[path] = handle
}

func (ht *handleTree) getHandle(path string) (handle Handler, params map[string]string, found bool) {
	for tmpl, hdl := range ht.mappings {
		pathSegs := strings.Split(strings.Trim(path, "/"), "/")
		tmplSegs := strings.Split(strings.Trim(tmpl, "/"), "/")

		if len(pathSegs) != len(tmplSegs) {
			continue
		}

		matched := true
		for i := 0; i < len(tmplSegs); i++ {
			if strings.HasPrefix(tmplSegs[i], ":") {
				if params == nil {
					params = make(map[string]string)
				}
				params[strings.TrimPrefix(tmplSegs[i], ":")] = pathSegs[i]
				continue
			}
			if pathSegs[i] != tmplSegs[i] {
				matched = false
				break
			}
		}

		if matched {
			handle, found = hdl, true
			return
		}

		params = nil
	}

	return
}

func (ht *handleTree) desc() {
	for path, _ := range ht.mappings {
		fmt.Printf("%s\t-->\t%s\n", ht.name, path)
	}
}

func main() {
	router := NewRouter()

	router.GET("/hello", func(c *Context) {
		c.w.WriteHeader(http.StatusOK)
		c.w.Write([]byte("hello, web framework"))
		return
	})

	router.GET("/user/:name", func(c *Context) {
		name := c.Param("name")
		if name == "" {
			name = "guest"
		}
		c.w.WriteHeader(http.StatusOK)
		c.w.Write([]byte(fmt.Sprintf("hello, %s", name)))
		return
	})

	router.POST("/panic", func(c *Context) {
		panic("intentionally")
	})

	err := router.Run(":8888")
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else {
		log.Println("error starting server")
	}
}
