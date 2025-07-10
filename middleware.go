package main

import (
	"fmt"
	"time"
)

type HandlerFunc func(w *ResponseWriter, r *Request)

type Middleware func(HandlerFunc) HandlerFunc

func Chain(mw []Middleware, final HandlerFunc) HandlerFunc {
	h := final

	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h

}
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w *ResponseWriter, r *Request) {
		start := time.Now()

		// actual handler
		next(w, r)

		fmt.Printf("%s %s -> %d (%s)\n",
			r.Method, r.Path,
			w.Status,
			time.Since(start),
		)
	}
}

func RecoverMiddleware(next HandlerFunc) HandlerFunc {
	return func(w *ResponseWriter, r *Request) {
		// The deferred function will execute after the surrounding function exits, no matter
		// how it exits (normally or via panic)
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Printf("panic in handler %s %s: %v\n", r.Method, r.Path, rec)

				w.WriteHeader(500)
				w.Header("Content-Type", "text/plain")
				w.Write([]byte("Internal Server Error"))
				w.Send()
			}
		}()

		next(w, r)
	}
}
