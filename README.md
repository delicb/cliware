# cliware
[![Go Report Card](https://goreportcard.com/badge/github.com/delicb/cliware)](https://goreportcard.com/report/github.com/delicb/cliware)
[![Build Status](https://travis-ci.org/delicb/cliware.svg?branch=master)](https://travis-ci.org/delicb/cliware)
[![codecov](https://codecov.io/gh/delicb/cliware/branch/master/graph/badge.svg)](https://codecov.io/gh/delicb/cliware)
![status](https://img.shields.io/badge/status-beta-red.svg)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/delicb/cliware)

Cliware is minimal HTTP client middleware library.

Main idea behind cliware is to make middleware concept usable for HTTP clients
as they are usable on servers. In order to achieve this, types and interfaces
that already exist in GoLang standard library, like 
[http.Handler](https://golang.org/pkg/net/http/#Handler), are defined for 
HTTP clients. 

## Install
Run `go get go.delic.rs/cliware` in terminal.

## Scope
Scope of this library is pretty small. It only defines required types (for
handler and middleware) and mechanism how they are chained. That is it.
No http client implementation (one will be release soon, but as separate 
project), no useful middlewares (also - there will be separate project).

## Dependencies
Cliware depends only on GoLang standard library. 
It requires GoLang 1.7+, because it uses `context` package. Earlier versions
might be possible and I might add support later.

## Name
Very creatively, name is combination of words CLI(ent) and (Middle)WARE. 

## Example
Since library only provides bases for other development, only few examples
will be shown. 

Middleware that would track execution time and log results of request sending
can look something like this.

```go
import (
	"context"
	"log"
	"net/http"
	"time"

	"go.delic.rs/cliware"
)

func RequestLog(name, value string) cliware.Middleware {
	return cliware.MiddlewareFunc(func(next cliware.Handler) cliware.Handler {
		return cliware.HandlerFunc(func(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
			// mark start time
			start := time.Now()

			// call next handler in chain
			resp, err = next.Handle(ctx, req)

			// mark end time and calculate duration
			end := time.Now()
			duration := end.Sub(start)

			// log what we gathered
			if err != nil {
				log.Fatalf("Request to \"%s\" took: %s and resulted in error: %s", req.URL, duration, err)
			} else {
				log.Printf("Request to \"%s\" took: %s, response code is: %d", req.URL, duration, resp.StatusCode)
			}

			// return results for further processing
			return resp, err
		})
	})
}
```

Previous example can seem pretty complicated at first. That is because we used
most complicated way to write middleware. This is most powerful way, but 
it requires a lot of code. In a lot of use cases, middleware only needs to
modify request or inspect response. For that purpose, there are RequestProcessor
and ResponseProcessor types. For example, if you want to modify request by
setting HTTP method to it, you can write simple middleware like this:

```go
func Method(method string) cliware.Middleware {
	return cliware.RequestProcessor(func(req *http.Request) error {
		req.Method = method
		return nil
	})
}
```
 
