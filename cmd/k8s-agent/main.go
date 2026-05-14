package main

import (
	"os"

	"github.com/threestoneliu/k8s-agent/cmd/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
