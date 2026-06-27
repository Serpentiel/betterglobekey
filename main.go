// Package main is the entry point for the betterglobekey application.
package main

import (
	"fmt"
	"os"

	"github.com/Serpentiel/betterglobekey/internal/cli"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
