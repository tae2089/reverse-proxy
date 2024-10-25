package server

import (
	"net/http"

	"github.com/tae2089/reverse-proxy/internal/server/controller"
	"github.com/tae2089/reverse-proxy/internal/server/middleware"
)

func newProxyRouter(router *http.ServeMux, proxyController controller.ProxyController, mode string, UrlPatternStr string) error {
	m := middleware.New(mode, UrlPatternStr)
	router.Handle("/", MultipleMiddleware(proxyController.ProxyRequestHandler(), m.GetMiddlewares()...))
	return nil
}

func newMetricRouter(router *http.ServeMux, proxyController controller.ProxyController) error {
	router.HandleFunc("/metrics", proxyController.Metrics())
	return nil
}

func MultipleMiddleware(h http.HandlerFunc, m ...middleware.MiddlewareFunc) http.HandlerFunc {
	if len(m) < 1 {
		return h
	}
	wrapped := h
	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}
	return wrapped

}
