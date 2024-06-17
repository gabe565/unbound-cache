package main

import (
	"os"

	"github.com/gabe565/unbound-cache/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
