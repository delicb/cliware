package cliware_test

import (
	"context"
	"fmt"
	"net/http"

	"net/url"

	c "go.delic.rs/cliware"
)

// HTTPError is struct holding information about HTTP response that went wrong.
type HTTPError struct {
	Status     string
	StatusCode int
	URL        string
	Method     string
}

// Error is implementation of error interface for HTTPError.
func (h *HTTPError) Error() string {
	return fmt.Sprintf("%s %s (%d %s)", h.Method, h.URL, h.StatusCode, h.Status)
}

// statusCodeToError is middleware that inspects HTTP response and if its
// response code is higher or equal to 400 creates HTTPError instance and
// passes it down the chain with response relevant response information.
func statusCodeToError() c.Middleware {
	return c.ResponseProcessor(func(resp *http.Response, err error) error {
		// no further processing if we already got error, but in different
		// middleware we can inspect error, replace it or suppress it
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return &HTTPError{
				Status:     resp.Status,
				StatusCode: resp.StatusCode,
				URL:        resp.Request.URL.String(),
				Method:     resp.Request.Method,
			}
		}
		return nil
	})
}

// notFoundResponse is helper method that always returns HTTP response with
// 404 Not Found status.
func notFoundResponse(ctx context.Context, req *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     http.StatusText(http.StatusNotFound),
		StatusCode: http.StatusNotFound,
		Request: &http.Request{
			URL: &url.URL{
				Path: "/some_path",
			},
			Method: "GET",
		},
	}, nil
}

func ExampleResponseProcessor() {
	_, err := statusCodeToError().Exec(c.HandlerFunc(notFoundResponse)).Handle(nil, nil)
	fmt.Println(err)
}
