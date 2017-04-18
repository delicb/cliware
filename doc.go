// Package cliware defines base type for HTTP client middleware support.
//
// Cliware does very little on its own. Main purpose is for other packages
// to implement useful HTTP client middlewares and to write clients using
// cliware and set of useful middlewares. That being said, cliware can be used
// on its own with GoLang standard library.
//
// Just as net/http package defines Handler interface that is used to write
// HTTP server handlers and compose them, this package defines similar concept
// but for HTTP client. net/http.Handler interface enabled a lot of third
// party middlewares to be written that are very useful. This package has the
// same idea. It provides type definition and some useful helper stuff around
// it (like ability to chain middlewares) in hope that other developer will
// catch up and write useful middlewares for everyone to use.
package cliware
