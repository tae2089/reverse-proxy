package options

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tae2089/reverse-proxy/internal/server"
)

type Options struct {
	Port            int
	MetricsPort     int
	ShutdownTimeOut int
	DisableMetrics  bool
	TargetHost      string
	Mode            string
	UrlPatternStr   string
	ApplicationName string
}

func New() *Options {
	return &Options{}
}

func (o *Options) Validate() error {
	if o.TargetHost == "" {
		return errors.New("target-host is required, please provide a target host. example: --target-host=http://localhost:8080")
	}
	return nil
}

func (o *Options) Complete(args []string, cmd *cobra.Command) error {
	o.Port = viper.GetInt("port")
	o.ShutdownTimeOut = viper.GetInt("shutdown-timeout")
	o.Mode = viper.GetString("mode")
	o.MetricsPort = viper.GetInt("metrics-port")
	o.TargetHost = viper.GetString("target-host")
	o.DisableMetrics = viper.GetBool("disable-metrics")
	o.UrlPatternStr = viper.GetString("url-patterns")
	o.ApplicationName = viper.GetString("application-name")
	return nil
}

func (o *Options) GetServerConfig() *server.Config {
	return &server.Config{
		Port:            o.Port,
		EnableMetrics:   !o.DisableMetrics,
		MetricsPort:     o.MetricsPort,
		TargetHost:      o.TargetHost,
		ShutdownTimeOut: time.Duration(o.ShutdownTimeOut) * time.Second,
		UrlPatternStr:   o.UrlPatternStr,
	}
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.Port, "port", 8080, "Port number to listen on, default is 8080 if not provided. example: --port=8080")
	cmd.Flags().IntVar(&o.MetricsPort, "metrics-port", 10250, "Port number to expose metrics, default is 10250 if not provided. example: --metrics-port=10250")
	cmd.Flags().IntVar(&o.ShutdownTimeOut, "shutdown-timeout", 30, "ShutDownTimeOut in seconds, default is 30 if not provided. example: --shutdown-timeout=10")
	cmd.Flags().StringVar(&o.TargetHost, "target-host", "", "Target host to proxy requests to, example: --target-host=http://localhost:8080")
	cmd.Flags().StringVar(&o.Mode, "mode", "otel", "Mode to run the server in, default is otel. example: --mode=otel")
	cmd.Flags().BoolVar(&o.DisableMetrics, "disable-metrics", false, "Disable metrics, default is false. example: --disable-metrics")
	cmd.Flags().StringVar(&o.UrlPatternStr, "url-patterns", "", "URL patterns to match. you can use pattern list separated by comma, e.g. /api,/api/{id},/api/v1/{id}")
	cmd.Flags().StringVar(&o.ApplicationName, "application-name", "demo", "Application name is target server name, default is demo. example: --application-name=demo")
}
