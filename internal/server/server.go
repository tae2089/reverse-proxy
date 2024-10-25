package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tae2089/reverse-proxy/internal/log"
	"github.com/tae2089/reverse-proxy/internal/observe"
	"github.com/tae2089/reverse-proxy/internal/server/controller"
	"golang.org/x/sync/errgroup"
)

const version = "dev"
const serviceName = "reverse-proxy"

type Config struct {
	EnableMetrics   bool
	Port            int
	MetricsPort     int
	TargetHost      string
	UrlPatternStr   string
	ApplicationName string
	ShutdownTimeOut time.Duration
}

type Server struct {
	ApplicationName string
	ProxyServer     *http.Server
	MetricsServer   *http.Server
	ShutdownTimeOut time.Duration
}

func (c *Config) Complete() (*Server, error) {
	proxyRouter := http.NewServeMux()
	metricsRouter := http.NewServeMux()
	// Create proxy,metrics handler
	proxyController := controller.New(c.TargetHost)
	if err := newProxyRouter(proxyRouter, proxyController, observe.OBSERVCE_MODE_OTEL, c.UrlPatternStr, c.EnableMetrics); err != nil {
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

	// Register observability
	observe.Register(s.ApplicationName, serviceName, version, observe.OBSERVCE_MODE_OTEL)

	// setup signal notify for graceful shutdown
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create errgroup
	g, gCtx := errgroup.WithContext(mainCtx)

	// Run servers
	s.runServers(g)
	log.Info("server started")

	// Graceful shutdown
	g.Go(func() error {
		<-gCtx.Done()
		return s.Stop()
	})

	// Wait for all servers to finish
	if err := g.Wait(); err != nil {
		return err
	}
	log.Info("shutting down")
	return nil
}

func (s *Server) runServers(g *errgroup.Group) error {
	// Run proxy server
	g.Go(func() error {
		if err := s.ProxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	// Run metrics server (if exists)
	if s.MetricsServer != nil {
		g.Go(func() error {
			if err := s.MetricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		})
	}
	return nil
}

func (s *Server) Stop() error {
	// Stop proxy server
	tctx := context.Background()
	// tctx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeOut)
	// defer cancel()
	var errWrap error = nil
	if err := s.ProxyServer.Shutdown(tctx); err != nil {
		errWrap = err
	}

	// Stop metrics server
	if s.MetricsServer != nil {
		if err := s.MetricsServer.Shutdown(tctx); err != nil {
			errWrap = errors.Join(errWrap, err)
		}
	}
	return errWrap
}
