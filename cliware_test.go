package cliware_test

import (
	"context"
	"net/http"
	"testing"

	"errors"

	"reflect"

	m "go.delic.rs/cliware"
)

func TestRequestHandlerFunc(t *testing.T) {
	var called bool
	handlerFunc := func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
		called = true
		return nil, nil
	}
	var handler m.Handler = m.HandlerFunc(handlerFunc)
	_, err := handler.Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !called {
		t.Error("Expected request handler function to be called.")
	}
}

func TestMiddlewareFunc(t *testing.T) {
	var called bool
	middlewareFunc := func(next m.Handler) m.Handler {
		called = true
		return nil
	}
	var middleware m.Middleware = m.MiddlewareFunc(middlewareFunc)
	middleware.Exec(nil)
	if !called {
		t.Error("Expected middleware func to be called.")
	}
}

func TestChainCreation(t *testing.T) {
	m1, _ := createMiddleware()
	m2, _ := createMiddleware()

	chain := m.NewChain(m1, m2)
	if len(chain.Middlewares()) != 2 {
		t.Error("Expected 2 middlewares in chain, found: ", len(chain.Middlewares()))
	}
}

func TestMiddlewareUse(t *testing.T) {
	m1, _ := createMiddleware()
	m2, _ := createMiddleware()
	chain := m.NewChain()

	chain.Use(m1)
	chain.Use(m2)
	if len(chain.Middlewares()) != 2 {
		t.Error("Expected 2 middlewares in chain, found: ", len(chain.Middlewares()))
	}
}

func TestMiddlewareUseFunc(t *testing.T) {
	chain := m.NewChain()
	chain.UseFunc(func(next m.Handler) m.Handler {
		return nil
	})
	if len(chain.Middlewares()) != 1 {
		t.Error("Expected 1 middleware in chain, found: ", len(chain.Middlewares()))
	}
}

func TestUseRequest(t *testing.T) {
	chain := m.NewChain()
	var called bool
	var validRequest bool
	templateReq, _ := http.NewRequest("GET", "http://localhost", nil)
	chain.UseRequest(func(req *http.Request) error {
		called = true
		validRequest = req == templateReq
		return nil
	})
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, templateReq)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !called {
		t.Error("Request middleware not called.")
	}
	if !*handlerCalled {
		t.Error("Final handler not called.")
	}
	if !validRequest {
		t.Error("Request handler did not receive expected request.")
	}
}

func TestUseResponse(t *testing.T) {
	chain := m.NewChain()
	var called bool
	chain.UseResponse(func(resp *http.Response, err error) error {
		called = true
		return nil
	})
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !called {
		t.Error("Response middleware not called.")
	}
	if !*handlerCalled {
		t.Error("Final handler not called.")
	}
}

func TestMiddlewareCalled(t *testing.T) {
	m1, m1Called := createMiddleware()
	m2, m2Called := createMiddleware()
	handler, handlerCalled := createHandler()
	chain := m.NewChain(m1, m2)
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !*m1Called {
		t.Error("m1 middleware not called.")
	}
	if !*m2Called {
		t.Error("m2 middleware not called.")
	}
	if !*handlerCalled {
		t.Error("Final handler not called.")
	}
}

func TestMiddlewareCalledWithParent(t *testing.T) {
	m1, m1Called := createMiddleware()
	m2, m2Called := createMiddleware()
	handler, handlerCalled := createHandler()

	chain := m.NewChain(m1)
	childChain := chain.ChildChain(m2)
	_, err := childChain.Exec(handler).Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !*m1Called {
		t.Error("m1 middleware not called.")
	}
	if !*m2Called {
		t.Error("m2 middleware not called.")
	}
	if !*handlerCalled {
		t.Error("Final handler not called.")
	}
}

func TestGetParent(t *testing.T) {
	chain := m.NewChain()
	childChain := chain.ChildChain()
	if childChain.Parent() != chain {
		t.Error("Parent middleware not set properly.")
	}
}

func TestRequestProcessorNoError(t *testing.T) {
	var processorCalled bool
	processor := m.RequestProcessor(func(req *http.Request) error {
		processorCalled = true
		return nil
	})
	chain := m.NewChain(processor)
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !processorCalled {
		t.Error("Request processor not called.")
	}
	if !*handlerCalled {
		t.Error("Handler was not called.")
	}
}

func TestRequestProcessorWithError(t *testing.T) {
	var processorCalled bool
	myErr := errors.New("custom error")
	processor := m.RequestProcessor(func(req *http.Request) error {
		processorCalled = true
		return myErr
	})
	chain := m.NewChain(processor)
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != myErr {
		t.Errorf("Expected error: \"%s\", got: \"%s\"", myErr, err)
	}
	if !processorCalled {
		t.Error("Request processor not called.")
	}
	if *handlerCalled {
		t.Error("Handler called even when middleware returned error.")
	}
}

func TestResponseProcessorNoError(t *testing.T) {
	var processorCalled bool
	processor := m.ResponseProcessor(func(resp *http.Response, err error) error {
		processorCalled = true
		return nil
	})
	chain := m.NewChain(processor)
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !processorCalled {
		t.Error("Response processor not called.")
	}
	if !*handlerCalled {
		t.Error("Handler was not called.")
	}
}

func TestResponseProcessorWithError(t *testing.T) {
	var processorCalled bool
	myErr := errors.New("custom error")
	processor := m.ResponseProcessor(func(resp *http.Response, err error) error {
		processorCalled = true
		return myErr
	})
	chain := m.NewChain(processor)
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != myErr {
		t.Errorf("Expected error: \"%s\", got: \"%s\"", myErr, err)
	}
	if !processorCalled {
		t.Error("Response processor not called.")
	}
	if !*handlerCalled {
		t.Error("Handler not called.")
	}
}

func TestContextProcessor_Exec(t *testing.T) {
	var processorCalled bool
	processor := m.ContextProcessor(func(ctx context.Context) context.Context {
		processorCalled = true
		return ctx
	})
	chain := m.NewChain(processor)
	handler, handlerCalled := createHandler()
	_, err := chain.Exec(handler).Handle(nil, nil)
	if err != nil {
		t.Error("Handle returned error: ", err)
	}
	if !processorCalled {
		t.Error("Context processor not called.")
	}
	if !*handlerCalled {
		t.Error("Handler was not called.")
	}
}

func TestCopy(t *testing.T) {
	processor := m.RequestProcessor(func(req *http.Request) error {
		return nil
	})
	originalProcessor := reflect.ValueOf(processor)
	chain := m.NewChain(processor).Copy()

	if len(chain.Middlewares()) != 1 {
		t.Fatal("Wrong number of middlewares in copied chain.")
	}
	copiedProcessor := reflect.ValueOf(chain.Middlewares()[0])
	if originalProcessor != copiedProcessor {
		t.Error("Got wrong middleware in copied chain.")
	}
}

func TestEmptyRequest(t *testing.T) {
	req := m.EmptyRequest()
	if req.Method != "GET" {
		t.Errorf("Empty request method wrong. Got: %s, expected: GET", req.Method)
	}
	if req.URL.Host != "" || req.URL.Scheme != "" || req.URL.Path != "" {
		t.Errorf("Empty request URL wrong. Got: %s, expected: <empty>", req.URL)
	}
	if req.Host != "" {
		t.Errorf("Empty request host wrong. Got %s, expected: <empty>", req.Host)
	}
	if req.ProtoMajor != 1 {
		t.Errorf("Empty request ProtoMajor wrong. Got: %d, exptected: 1", req.ProtoMajor)
	}
	if req.ProtoMinor != 1 {
		t.Errorf("Empty request ProtoMinor wrong. Got: %d, exptected: 1", req.ProtoMajor)
	}
	if req.Proto != "HTTP/1.1" {
		t.Errorf("Empty request Proto wrong. Got: %s, exptected HTTP/1.1", req.Proto)
	}
}

func createMiddleware() (middleware m.Middleware, called *bool) {
	var middlewareCalled bool
	middleware = m.MiddlewareFunc(func(next m.Handler) m.Handler {
		middlewareCalled = true
		return m.HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
			return next.Handle(ctx, req)
		})
	})
	return middleware, &middlewareCalled
}

func createHandler() (handler m.Handler, called *bool) {
	var handlerCalled bool
	handler = m.HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
		handlerCalled = true
		return nil, nil
	})
	return handler, &handlerCalled
}
