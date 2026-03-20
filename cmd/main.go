package main

import (
	"os"

	"github.com/somaz94/bash-pilot/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
