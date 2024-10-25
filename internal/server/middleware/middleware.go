package middleware

import "net/http"

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

type Middleware interface {
	GetMiddlewares() []MiddlewareFunc
}

// New creates a new middleware
func New(mode, UrlPatternStr string, enableMetrics bool) Middleware {
	var m Middleware
	switch mode {
	case "otel":
		m = newOtelMiddleware(enableMetrics, UrlPatternStr)
	default:
		m = newOtelMiddleware(enableMetrics, UrlPatternStr)
	}
	return m
}
