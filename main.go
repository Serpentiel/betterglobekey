// Package main is the entry point for the betterglobekey application.
package main

import (
	"fmt"
	"os"

	"github.com/Serpentiel/betterglobekey/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
