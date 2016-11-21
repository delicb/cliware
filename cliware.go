package cliware

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

///////////////////////////////////////////////////////////////////////////////
// Request handler mechanism
///////////////////////////////////////////////////////////////////////////////

// Handler is interface that defines processing HTTP client request.
type Handler interface {
	// Handle executes HTTP request based on provided parameters.
	// Depending on implementation, provided context might or not be used.
	// However, if handler really performs HTTP request (as opposed to doing
	// some middleware work, like logging or similar), provided context SHOULD
	// be used on request as intended by http.Request.WithRequest method.
	Handle(ctx context.Context, req *http.Request) (resp *http.Response, err error)
}

// HandlerFunc is function variant of RequestHandler interface.
type HandlerFunc func(ctx context.Context, req *http.Request) (resp *http.Response, err error)

// Handle is implementation of RequestHandler interface
func (rhf HandlerFunc) Handle(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	return rhf(ctx, req)
}

///////////////////////////////////////////////////////////////////////////////
// Middleware mechanism definition
///////////////////////////////////////////////////////////////////////////////

// Middleware defines interface for chaining Handlers.
type Middleware interface {
	// Exec returns Handler that does some useful work. Implementation SHOULD
	// call provided next Handler somewhere in returned Handler to ensure chain
	// is not broken. However, implementation might choose not to call next
	// handler based on external influence, error checking, etc. In that case,
	// returned Handler MUST return error.
	Exec(next Handler) Handler
}

// MiddlewareFunc is function variant of Middleware interface.
type MiddlewareFunc func(handler Handler) Handler

// Exec is implementation of Middleware interface.
func (mf MiddlewareFunc) Exec(handler Handler) Handler {
	return mf(handler)
}

// RequestProcessor is function for modification of HTTP request.
// It is intended as form of simple Middleware for middlewares that only need
// to change request that is being sent. Provided request can be modified
// as required. Returned error (if any) will stop middleware chain execution
// and same error will be returned to caller.
type RequestProcessor func(req *http.Request) error

// Exec is implementation of Middleware interface.
func (rp RequestProcessor) Exec(handler Handler) Handler {
	return HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
		err = rp(req)
		if err != nil {
			return nil, err
		}

		resp, err = handler.Handle(ctx, req)
		return resp, err
	})
}

// ResponseProcessor is function for inspection of HTTP response.
// It is intended as form of a simple Middleware for middlewares that only
// need to inspect responses. E.g. they can log some information or inspect
// response to determine if error occurred. Provided response if one obtained
// from sending HTTP request and should not be modified. Provided error is
// original error returned after request or error returned by previous
// middleware. If middleware wants to change error - it should return it.
// Otherwise, if there is not error or existing (provided) error should not be
// changed, middleware should return nil.
type ResponseProcessor func(resp *http.Response, err error) error

// Exec is implementation of Middleware interface.
func (rp ResponseProcessor) Exec(handler Handler) Handler {
	return HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
		resp, err = handler.Handle(ctx, req)
		newErr := rp(resp, err)
		if newErr != nil {
			return resp, newErr
		}
		return resp, err
	})
}

// ContextProcessor is function for managing request context.
// It is intended as for of simple middleware for middlewares that only
// need to modify context before sending request.
type ContextProcessor func(ctx context.Context) context.Context

// Exec is implementation of Middleware interface.
func (cp ContextProcessor) Exec(handler Handler) Handler {
	return HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
		ctx = cp(ctx)
		return handler.Handle(ctx, req)
	})
}

// Chain is Middleware implementation capable of executing multiple
// middlewares. On top of that, chain is aware of its parent middleware and
// executes it during its own execution.
type Chain struct {
	middlewares []Middleware
	parent      Middleware
}

// NewChain creates and returns middleware chain with provided middlewares
// and no parent. If you need to create new chain with some parent, use
// ChildChain method.
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
		parent:      nil,
	}
}

// Copy creates new chain with all middlewares copied to it.
func (c *Chain) Copy() *Chain {
	middlewareCopy := make([]Middleware, len(c.middlewares))
	copy(middlewareCopy, c.middlewares)
	return &Chain{
		middlewares: middlewareCopy,
		parent:      nil,
	}
}

// ChildChain creates new Middleware chain with current chain as parent.
func (c *Chain) ChildChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
		parent:      c,
	}
}

// Middlewares returns all middlewares for this chain. Parent middlewares
// not included.
func (c *Chain) Middlewares() []Middleware {
	return c.middlewares
}

// Parent returns parent middleware of this chain.
func (c *Chain) Parent() Middleware {
	return c.parent
}

// Exec is implementation of Middleware interface that executes all middlewares
// in chain, including parent middleware.
func (c *Chain) Exec(handler Handler) Handler {
	finalHandler := handler

	// Make sure to run own middlewares first... Because of the way middlewares
	// are composed, ones called first will override ones called later and
	// we want to be able to override middlewares in child chain.
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		finalHandler = c.middlewares[i].Exec(finalHandler)
	}

	// if we have parent, make sure to call it too...
	if c.parent != nil {
		finalHandler = c.parent.Exec(finalHandler)
	}

	return finalHandler
}

// Use adds provided middleware to current middleware chain.
func (c *Chain) Use(m Middleware) {
	c.middlewares = append(c.middlewares, m)
}

// UseAll adds all provided middlewares to current middleware chain.
func (c *Chain) UseAll(m ...Middleware) {
	c.middlewares = append(c.middlewares, m...)
}

// UseFunc adds provided function to current middleware chain.
func (c *Chain) UseFunc(m func(handler Handler) Handler) {
	c.middlewares = append(c.middlewares, MiddlewareFunc(m))
}

// UseRequest adds provided function as request middleware.
func (c *Chain) UseRequest(m func(req *http.Request) error) {
	c.Use(RequestProcessor(m))
}

// UseResponse add provided function as response middleware.
func (c *Chain) UseResponse(m func(resp *http.Response, err error) error) {
	c.Use(ResponseProcessor(m))
}

// EmptyRequest creates new empty instance of *http.Request.
// It is good starting point for initial request instance for middleware chain.
// In contrast to http.NewRequest, this function does not require any parameters.
// Any value can be overridden by middlewares. Request method is set to GET,
// just because it is sane default.
func EmptyRequest() *http.Request {
	req := &http.Request{
		Method:     "GET",
		URL:        &url.URL{},
		Host:       "",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Proto:      "HTTP/1.1",
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
	}
	return req
}
