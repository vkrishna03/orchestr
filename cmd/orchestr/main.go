package main

import (
	"os"

	"github.com/vkrishna03/orchestr/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
