// Package app wires the domain and infrastructure together into a runnable daemon.
package app

import (
	"context"

	"github.com/Serpentiel/betterglobekey/internal/domain/switcher"
)

// Presser handles a single Globe key press.
type Presser interface {
	// Press reacts to a Globe key release.
	Press()
}

// Keyboard is the port that delivers Globe key releases.
type Keyboard interface {
	// Listen blocks, invoking onKeyUp for each Globe key release, until Stop is called.
	Listen(onKeyUp func())
	// Stop unblocks Listen and releases the keyboard hook.
	Stop()
}

// Logger is the logging port used by the daemon.
type Logger interface {
	switcher.Logger
	// Close flushes any buffered log entries.
	Close() error
}

// Daemon listens for Globe key presses and drives the switcher in response.
type Daemon struct {
	switcher Presser
	keyboard Keyboard
	logger   Logger
}

// NewDaemon assembles a Daemon from its collaborators.
func NewDaemon(s Presser, keyboard Keyboard, logger Logger) *Daemon {
	return &Daemon{switcher: s, keyboard: keyboard, logger: logger}
}

// Run starts the keyboard listener and blocks until ctx is cancelled, then shuts
// down cleanly.
func (d *Daemon) Run(ctx context.Context) error {
	defer func() { _ = d.logger.Close() }()

	go d.keyboard.Listen(d.switcher.Press)

	d.logger.Info("event handler has been initialized")

	<-ctx.Done()

	d.keyboard.Stop()
	d.logger.Info("shutting down")

	return nil
}
