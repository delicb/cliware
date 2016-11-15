package cliware_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"context"

	"go.delic.rs/cliware"
)

func Example() {
	// start test http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// print some headers
		fmt.Println("User-Agent: ", r.Header.Get("User-Agent"))
		fmt.Println("Content-Type: ", r.Header.Get("Content-Type"))
		fmt.Println("Custom-Header: ", r.Header.Get("X-Custom-Header"))

		// dump body to stdout
		io.Copy(os.Stdout, r.Body)
		defer r.Body.Close()

		// return some response
		w.WriteHeader(200)
		w.Write([]byte("My shiny server response"))
	}))

	// define middleware that will set URL to request - endpoint to send request to
	urlMiddleware := cliware.RequestProcessor(func(req *http.Request) error {
		u, err := url.Parse(server.URL)
		if err != nil {
			return err
		}
		req.URL = u
		return nil
	})

	// middleware that our headers to request
	headersMiddleware := cliware.RequestProcessor(func(req *http.Request) error {
		req.Header.Set("User-Agent", "Cliware")
		req.Header.Set("X-Custom-Header", "whatever")
		return nil
	})

	// middleware for setting body to request. Also, it sets Content-Type header.
	bodyMiddleware := cliware.RequestProcessor(func(req *http.Request) error {
		req.Body = ioutil.NopCloser(strings.NewReader("request data"))
		req.Header.Set("Content-Type", "application/text")
		return nil
	})

	// middleware for checking for error and printing response to StdOut
	responseMiddleware := cliware.ResponseProcessor(func(resp *http.Response, err error) error {
		if err != nil {
			fmt.Println("Got error:", err)
			return err
		}
		defer resp.Body.Close()
		fmt.Println()
		fmt.Println("Got response:")
		io.Copy(os.Stdout, resp.Body)
		return nil
	})

	// example of more complex middleware that can track entire request
	traceMiddleware := cliware.MiddlewareFunc(func(next cliware.Handler) cliware.Handler {
		return cliware.HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
			// do anything before request
			fmt.Printf("\n*** Before sending request.\n")

			// call next middleware
			resp, err = next.Handle(ctx, req)

			// do anything after request
			fmt.Printf("\n*** After sending request.\n")

			// return result
			return resp, err
		})
	})

	// final handler - one that will really send request
	finalHandler := cliware.HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
		r := req.WithContext(ctx)
		return http.DefaultClient.Do(r)
	})

	// middlewares will be applied in oder defined here
	chain := cliware.NewChain(urlMiddleware, headersMiddleware, bodyMiddleware, responseMiddleware)
	// other way to add middleware is by using Use* method
	chain.Use(traceMiddleware)
	// execute chain and final middleware
	chain.Exec(finalHandler).Handle(context.Background(), cliware.EmptyRequest())

	server.Close()

	// Output:
	// *** Before sending request.
	// User-Agent:  Cliware
	// Content-Type:  application/text
	// Custom-Header:  whatever
	// request data
	// *** After sending request.

	// Got response:
	// My shiny server response
}
