# cliware
[![Go Report Card](https://goreportcard.com/badge/github.com/delicb/cliware)](https://goreportcard.com/report/github.com/delicb/cliware)
[![Build Status](https://travis-ci.org/delicb/cliware.svg?branch=master)](https://travis-ci.org/delicb/cliware)
[![codecov](https://codecov.io/gh/delicb/cliware/branch/master/graph/badge.svg)](https://codecov.io/gh/delicb/cliware)
![status](https://img.shields.io/badge/status-stable-brightgreen.svg)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/delicb/cliware)

Cliware - middlewares from clients.

Cliware is library that defines concept of middlewares for HTTP clients. In 
Go community, middlewares are known concept for server side development. They
are widely used and there are a lot of useful middleware implementations around.

When it comes to implementing HTTP client, there is no standard. There are some
great libraries, but there is no real plug-and-play mechanism as there is for
common server handlers. Idea behind this library is to make it possible for 
others to write client middlewares and compose them to create HTTP clients.

## Install
Run `go get go.delic.rs/cliware` in terminal.

## Usage
Two main types in `cliware` are `Handler` and `Middleware`. They are both 
interfaces with some helper around them. 

### Handler
`Handler` defines `Handle` method whose job is to take context in which it is
being executed and instance of `*http.Request` and return `*http.Response` (
with optional error). That is it. It is pretty similar to what `*http.Client.Do`
method does, except it does not accept context as first parameter.

At least one Handler has to exist and that is one that will send request over
network. Simple implementation can look like this:
```go
func finalHandler(ctx *context.Context, req *http.Request) (resp *http.Response, err error) {
    return http.DefaultClient.Do(req.WithContext(ctx))
}
```

### Middleware
`Middleware` is interface that also defines only one method. That method is
`Exec` and it accepts next `Handler` in to be called and returns new `Handler`.

Middlewares can wrap handlers based on some parameters, do some stuff before and
after next handler, etc, modify request, inspect response, etc. Middleware HAS TO
call next handler that was provided to `Exec` or else entire chain will stop
executing.

### HandlerFunc
`HandlerFunc` is a function type that has same signature as `Handle` method
from `Handler` interface. You can see it as utility that will convert your 
function with proper signature into `Handler` implementation.

### MiddlewareFunc 
`MiddlewareFunc` is same principle from `HandlerFunc`, just applied to `Middleware`.

### RequestProcessor
`RequestProcessor` is function type that implements `Middleware` interface. It is
not special type of `Middleware`, but just convenience for middlewares that only
need to do something with request and do not care about response. This allows for
such middlewares to have less verbose signature and it automatically calls next
handler.

### ResponseProcessor
`ResponseProcessor` is same principle as `RequestProcessor`, just applied to
responses. So, if middleware only needs to inspect response you can use this
convenience function.

### ContextProcessor
`ContextProcessor` is same principle as `RequestProcessor` and `ResponseProcessor`
but applied only to `context.Context` that is passed as first parameter to each
handler. It can be used if middleware only needs to modify context (like setting
timeout or deadline) to context.

### Chain
`Chain` is special kind of `Middleware`. Instead of wrapping one `Handler` with
another, `Chain` can hold multiple middlewares that will all be applied in 
sequence, thus effectively wrapping final `Handler`.

That is all chain does (plus some additional utility methods for adding other
middlewares to the chain).

## Scope
Scope of this library is pretty small. It only defines required types (for
handler and middleware) and mechanism how they are chained. That is it.
No useful middlewares (I am writing library of them, still work in progress but 
check out [cliware-middlewares](https://github.com/delicb/cliware-middlewares)),
No http client implementation (also writing one, check out 
[GWC](https://github.com/delicb/gwc)). 

## Dependencies
Cliware depends only on GoLang standard library. 
It requires GoLang 1.7+, because it uses `context` package. Earlier versions
might be possible and I might add support later.

## Contributing
Most appreciated contributions would be written middlewares that use Cliware.
If you do write some, please ping me to add link to this description.

Of course, feel free to open tickets (or even better, pull requests). I am opened
to suggestions. 

## Name
Very creatively, name is combination of words CLI(ent) and (Middle)WARE. 

## Examples
There are (testable) examples in repository itself. Just checkout files whose
name starts with `example_`

## Licence
Cliware is released under MIT licence.
