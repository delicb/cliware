package cliware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/delicb/cliware"
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

func Query(method string) cliware.Middleware {
	return cliware.RequestProcessor(func(req *http.Request) error {
		req.Method = method
		return nil
	})
}
