package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in the order they are passed (first middleware wraps all others)
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middlewares in reverse order so the first middleware is the outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
