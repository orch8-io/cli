package main

import (
	"os"

	"github.com/orch8-io/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
