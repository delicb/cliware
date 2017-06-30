package cliware_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	c "github.com/delicb/cliware"
)

var server *httptest.Server

// startServer starts dummy example server that will be used to demonstrate
// how middlewares work.
func startServer() {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// print some headers
		fmt.Println("User-Agent: ", r.Header.Get("User-Agent"))
		fmt.Println("Custom-Header: ", r.Header.Get("X-Custom-Header"))

		// dump body to stdout
		_, err := io.Copy(os.Stdout, r.Body)
		if err != nil {
			panic(err)
		}
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()

		// return some response
		w.WriteHeader(200)
		_, err = w.Write([]byte("My shiny server response"))
		if err != nil {
			panic(err)
		}
	}))
}

// stopServer stops testing server
func stopServer() {
	server.Close()
}

// serverURL sets request URL to match URL of provided test server.
// This is example of middleware that modifies request before sending it.
// It uses utility function RequestProcessor.
func serverURL(server *httptest.Server) c.Middleware {
	return c.RequestProcessor(func(r *http.Request) error {
		u, err := url.Parse(server.URL)
		if err != nil {
			return err
		}
		r.URL = u
		return nil
	})
}

// header adds header with provided name and value to request.
// This is another example of middleware that modifies request before sending
// it.
func header(name, value string) c.Middleware {
	return c.RequestProcessor(func(r *http.Request) error {
		r.Header.Set(name, value)
		return nil
	})
}

// bodyToStdout reads response body and copies it to standard output.
// This is example of middleware that inspects response and does something
// with returned body.
func bodyToStdout() c.Middleware {
	return c.ResponseProcessor(func(resp *http.Response, err error) error {
		if err != nil {
			log.Println("Got error:", err)
			return err
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		fmt.Println()
		fmt.Println("Got response:")
		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	})
}

// trace writes something to standard output before and after request has been sent.
// This is example of middleware without utility functions that does something
// both before and after request has been sent.
func trace(next c.Handler) c.Handler {
	return c.HandlerFunc(func(req *http.Request) (resp *http.Response, err error) {
		// do anything before request
		fmt.Println("*** Before sending request.")

		// call next middleware
		resp, err = next.Handle(req)

		// do anything after request
		fmt.Println("*** After sending request.")

		// return result
		return resp, err
	})
}

// sender does actual request sending using htt.Client.
// this is final handler that has to be called and it has to be passed to
// chain Exec method.
func sender(req *http.Request) (resp *http.Response, err error) {
	return http.DefaultClient.Do(req)
}

func Example() {
	startServer()
	defer stopServer()

	// middlewares will be applied in order defined here
	chain := c.NewChain(
		serverURL(server),
		header("User-Agent", "Cliware"),
		header("X-Custom-Header", "whatever"),
		bodyToStdout(),
	)
	// other way to add middlewares is by using Use* method
	chain.UseFunc(trace)
	// execute chain and final middleware
	_, err := chain.Exec(c.HandlerFunc(sender)).Handle(c.EmptyRequest())
	if err != nil {
		panic(err)
	}
	// Output:
	// *** Before sending request.
	// User-Agent:  Cliware
	// Custom-Header:  whatever
	// *** After sending request.
	//
	// Got response:
	// My shiny server response
}
