package main

import (
	"os"
	"runtime"

	"github.com/tae2089/reverse-proxy/cmd"
)

func main() {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	cmd := cmd.NewReverseProxyCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
