package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, mw := range middlewares {
		h = mw(h)
	}

	return h
}
