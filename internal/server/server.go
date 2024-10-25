package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tae2089/reverse-proxy/internal/log"
	"github.com/tae2089/reverse-proxy/internal/observe"
	"github.com/tae2089/reverse-proxy/internal/server/controller"
)

const version = "dev"
const serviceName = "reverse-proxy"

type Config struct {
	Port            int
	EnableMetrics   bool
	MetricsPort     int
	TargetHost      string
	ShutdownTimeOut time.Duration
	UrlPatternStr   string
	ApplicationName string
}

type Server struct {
	ProxyServer     *http.Server
	MetricsServer   *http.Server
	ShutdownTimeOut time.Duration
	ApplicationName string
}

func (c *Config) Complete() (*Server, error) {
	proxyRouter := http.NewServeMux()
	metricsRouter := http.NewServeMux()
	// Create proxy,metrics handler
	proxyController := controller.New(c.TargetHost)
	if err := newProxyRouter(proxyRouter, proxyController, observe.OBSERVCE_MODE_OTEL, c.UrlPatternStr); err != nil {
		return nil, err
	}
	if err := newMetricRouter(metricsRouter, proxyController); err != nil {
		return nil, err
	}

	// Create server
	svr := &Server{
		ProxyServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", c.Port),
			Handler: proxyRouter,
		},
		MetricsServer:   nil,
		ShutdownTimeOut: c.ShutdownTimeOut,
	}

	// Enable metrics server
	if c.EnableMetrics {
		svr.MetricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", c.MetricsPort),
			Handler: metricsRouter,
		}
	}
	return svr, nil
}

func (s *Server) Run() error {
	observe.Register(s.ApplicationName, serviceName, version, observe.OBSERVCE_MODE_OTEL)
	// Run proxy server
	s.runServers()
	log.Info("server started")
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeOut)
	defer cancel()

	if err := s.Stop(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	log.Info("shutting down")
	return nil
}

func (s *Server) runServers() error {
	// Run proxy server
	go func() {
		if err := s.ProxyServer.ListenAndServe(); err != nil {
			return
		}
	}()

	// Run metrics server (if exists)
	if s.MetricsServer != nil {
		go func() {
			if err := s.MetricsServer.ListenAndServe(); err != nil {
				return
			}
		}()
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	// Stop proxy server
	if err := s.ProxyServer.Shutdown(ctx); err != nil {
		return err
	}
	// Stop metrics server
	if s.MetricsServer != nil {
		if err := s.MetricsServer.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
