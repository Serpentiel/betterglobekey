// Package main is the entry point for the betterglobekey application.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Serpentiel/betterglobekey/internal/cli"
)

// version and commit are set at build time via
// -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = ""
)

// init pins the main goroutine to the main OS thread so the HUD's AppKit run
// loop runs on the main thread, as macOS requires.
func init() {
	runtime.LockOSThread()
}

func main() {
	if err := cli.Execute(version, commit); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
