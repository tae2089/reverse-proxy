package controller

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ProxyController interface {
	Metrics() func(http.ResponseWriter, *http.Request)
	ProxyRequestHandler() http.HandlerFunc
}

func New(targetHost string) ProxyController {
	proxy := newProxy(targetHost)
	ctrl := &proxyController{
		reverseProxy: proxy,
	}
	return ctrl
}

type proxyController struct {
	reverseProxy *httputil.ReverseProxy
}

func (p *proxyController) ProxyRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p.reverseProxy.ServeHTTP(w, r)
	}
}

func (p *proxyController) Metrics() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		promhttp.Handler().ServeHTTP(w, r)
	}
}

func newProxy(targetHost string) *httputil.ReverseProxy {
	url, err := url.Parse(targetHost)
	if err != nil {
		panic(err)
	}
	return httputil.NewSingleHostReverseProxy(url)
}
