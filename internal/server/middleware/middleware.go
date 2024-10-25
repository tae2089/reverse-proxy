package middleware

import "net/http"

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

type Middleware interface {
	GetMiddlewares() []MiddlewareFunc
}

// New creates a new middleware
func New(mode, UrlPatternStr string) Middleware {
	var m Middleware
	switch mode {
	case "otel":
		m = newOtelMiddleware(UrlPatternStr)
	default:
		m = newOtelMiddleware(UrlPatternStr)
	}
	return m
}
