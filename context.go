package cliware

import (
	"context"
	"net/http"
)

type clientKeyType string

var clientKey clientKeyType = "client"

// SetClient adds client instance to provided context and returns new
// context with client value.
func SetClient(ctx context.Context, client *http.Client) context.Context {
	c := context.WithValue(ctx, clientKey, client)
	return c
}

// GetClient returns client instance from context. If client is not set, nil
// will be returned.
func GetClient(ctx context.Context) *http.Client {
	cl := ctx.Value(clientKey)
	if cl == nil {
		return nil
	}
	if client, ok := cl.(*http.Client); ok {
		return client
	}
	return nil
}
