package cliware_test

import (
	"context"
	"fmt"
	"net/http"

	c "github.com/delicb/cliware"
)

// basicAuth is simple middleware that modifies request by adding basic authentication
// to it.
func basicAuth(username, password string) c.Middleware {
	return c.RequestProcessor(func(req *http.Request) error {
		req.SetBasicAuth(username, password)
		return nil
	})
}

func nilHandler(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}

func ExampleRequestProcessor() {
	req := c.EmptyRequest()
	basicAuth("user", "pass").Exec(c.HandlerFunc(nilHandler)).Handle(nil, req)
	username, password, ok := req.BasicAuth()
	fmt.Println(ok)
	fmt.Println(username)
	fmt.Println(password)
	// Output:
	// true
	// user
	// pass
}
