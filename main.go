// Package main is the entry point for the betterglobekey application.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Serpentiel/betterglobekey/internal/cli"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

// init pins the main goroutine to the main OS thread so the HUD's AppKit run
// loop runs on the main thread, as macOS requires.
func init() {
	runtime.LockOSThread()
}

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
