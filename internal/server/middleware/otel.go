package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tae2089/reverse-proxy/internal/log"
	"github.com/tae2089/reverse-proxy/internal/server/domain"
	"github.com/tae2089/reverse-proxy/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type otelMiddleware struct {
	IsEnabledMeasureLatency bool
	Gauge                   prometheus.Gauge
	Tracer                  trace.Tracer
	Props                   propagation.TextMapPropagator
	httpLatencyHistogram    *prometheus.HistogramVec
	PatternTree             *utils.Tree
	HttpRequestsCounter     *prometheus.CounterVec
}

// GetMiddlewares implements Middleware.
func (m *otelMiddleware) GetMiddlewares() []MiddlewareFunc {
	return []MiddlewareFunc{
		m.SetUpMiddleware,
		m.LoggingMiddleware,
		m.MetricsMiddleware,
		m.TraceIDMiddleware,
		m.TimerMiddleware,
	}
}

func (m *otelMiddleware) SetUpMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new ResponseCapture
		rec := domain.NewResponseCapture(w)
		ctx := context.WithValue(r.Context(), "rec", rec)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

// LoggingMiddleware logs the start and end of a request
func (m *otelMiddleware) LoggingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get ResponseCapture from context
		rec := r.Context().Value("rec").(*domain.ResponseCapture)
		// Call the next handler
		h.ServeHTTP(rec, r)
		// Capture the response
		response := rec.ToHttpResponse(r)
		// Get duration from context
		duration := r.Context().Value("latency").(time.Duration)
		//logging
		loggingFields := log.GetLoggingFieldsByRequest(r, response.StatusCode, duration)
		log.Info("service call", loggingFields...)
	}
}

func (m *otelMiddleware) TimerMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start the timer
		startTime := time.Now()
		// Call the next handler
		h.ServeHTTP(w, r)
		//calculate the latency
		latency := time.Since(startTime)
		// save the latency in the context
		ctx := context.WithValue(r.Context(), "latency", latency)
		// update the request with the new context
		*r = *r.WithContext(ctx)
	}
}

// TotalConnectionsMiddleware increments the total connections counter
func (m *otelMiddleware) MetricsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the measure latency is disabled, skip the measure
		if !m.IsEnabledMeasureLatency {
			h.ServeHTTP(w, r)
		}
		// Increment the total connections counter
		m.Gauge.Inc()
		h.ServeHTTP(w, r)
		m.Gauge.Dec()
		// Get ResponseCapture from context
		rec := r.Context().Value("rec").(*domain.ResponseCapture)
		response := rec.ToHttpResponse(r)
		// Get duration from context
		duration := r.Context().Value("latency").(time.Duration)
		m.measureRequest(r.URL.Path, r.Method, duration, response.StatusCode)
	}
}

// TraceIDMiddleware adds a trace ID to the request context
func (m *otelMiddleware) TraceIDMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//if the request has a traceparent header, extract the trace ID
		// if not, create a new trace ID
		var targetCtx context.Context
		if r.Header.Get("Traceparent") != "" {
			log.Info("Get header", zap.Any("Header", r.Header))
			targetCtx = m.Props.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		} else {
			targetCtx = r.Context()
		}
		// Create a new span with the trace ID
		// and inject the span context into the request headers
		ctx, span := m.Tracer.Start(targetCtx, "reverse-proxy")
		m.Props.Inject(ctx, propagation.HeaderCarrier(r.Header))
		// Update the request with the new context
		*r = *r.WithContext(ctx)
		h.ServeHTTP(w, r)
		span.End()
	}
}

func (m *otelMiddleware) measureRequest(path string, method string, duration time.Duration, statusCode int) {
	var replacedPath string = m.measureLatency(path, method, duration)
	m.countStatusCode(replacedPath, method, statusCode)
}

func (m *otelMiddleware) measureLatency(path string, method string, duration time.Duration) string {
	var pathPattern string = m.PatternTree.Search(path)
	var replacedPath string = m.PatternTree.ReplaceWithPattern(path, pathPattern)
	m.httpLatencyHistogram.WithLabelValues(replacedPath, method).Observe(duration.Seconds())
	return replacedPath
}

func (m *otelMiddleware) countStatusCode(path, method string, statusCode int) {
	var status string
	switch {
	case statusCode >= 200 && statusCode < 300:
		status = "2xx"
	case statusCode >= 300 && statusCode < 400:
		status = "3xx"
	case statusCode >= 400 && statusCode < 500:
		status = "4xx"
	case statusCode >= 500 && statusCode < 600:
		status = "5xx"
	default:
		status = "xxx"
	}
	m.HttpRequestsCounter.WithLabelValues(path, method, status).Inc()
}

func newOtelMiddleware(enableMetrics bool, UrlPatternStr string) Middleware {
	tracer := otel.GetTracerProvider().Tracer("reverse-proxy")
	var gauge prometheus.Gauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "total_connections",
		Help: "Total connections",
	})

	// Define a new Histogram metric
	var httpLatencyHistogram *prometheus.HistogramVec = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_latency",
			Help:    "Latency of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	var httpRequestsCounter *prometheus.CounterVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests",
			Help: "Total counte of HTTP requests by status code, path and method",
		},
		[]string{"path", "method", "status_code"},
	)

	pattenrTree := utils.NewTree()
	urlPatterns := strings.Split(UrlPatternStr, ",")
	for _, pattern := range urlPatterns {
		pattenrTree.Insert(pattern)
	}

	m := &otelMiddleware{
		Gauge:                   gauge,
		Tracer:                  tracer,
		Props:                   otel.GetTextMapPropagator(),
		httpLatencyHistogram:    httpLatencyHistogram,
		PatternTree:             pattenrTree,
		HttpRequestsCounter:     httpRequestsCounter,
		IsEnabledMeasureLatency: enableMetrics,
	}
	return m
}
