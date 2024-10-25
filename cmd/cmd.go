package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tae2089/reverse-proxy/cmd/options"
)

func NewReverseProxyCommand() *cobra.Command {
	opts := options.New()
	cmd := &cobra.Command{
		Short: "Launch reverse-proxy-server",
		Long:  "Launch reverse-proxy-server",
		Use:   "reverse-proxy-server [flags]",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete options
			if err := opts.Complete(args, cmd); err != nil {
				return err
			}
			// Validate options
			if err := opts.Validate(); err != nil {
				return err
			}
			// Get server Config
			serverConfig := opts.GetServerConfig()
			// Run server
			svr, err := serverConfig.Complete()
			if err != nil {
				return err
			}
			return svr.Run()
		},
	}
	opts.AddFlags(cmd)
	return cmd
}
