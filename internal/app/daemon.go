// Package app wires the domain and infrastructure together into a runnable daemon.
package app

import (
	"context"
	"sync"

	"github.com/Serpentiel/betterglobekey/internal/domain/switcher"
)

// Presser handles a single Globe key press. The reverse flag inverts the
// cycling direction (e.g. when a modifier is held).
type Presser interface {
	// Press reacts to a Globe key release.
	Press(reverse bool)
}

// Keyboard is the port that delivers Globe key releases.
type Keyboard interface {
	// Listen blocks, invoking onPress for each Globe key release (with reverse set
	// when a modifier is held), until Stop is called.
	Listen(onPress func(reverse bool))
	// Stop unblocks Listen and releases the keyboard hook.
	Stop()
}

// Logger is the logging port used by the daemon.
type Logger interface {
	switcher.Logger
	// Close flushes any buffered log entries.
	Close() error
}

// Daemon listens for Globe key presses and drives the switcher in response. The
// switcher can be swapped at runtime via Reload (used for live config reload).
type Daemon struct {
	mu       sync.RWMutex
	switcher Presser
	keyboard Keyboard
	logger   Logger
}

// NewDaemon assembles a Daemon from its collaborators.
func NewDaemon(s Presser, keyboard Keyboard, logger Logger) *Daemon {
	return &Daemon{switcher: s, keyboard: keyboard, logger: logger}
}

// Reload atomically swaps the active switcher, e.g. after the configuration changes.
func (d *Daemon) Reload(s Presser) {
	d.mu.Lock()
	d.switcher = s
	d.mu.Unlock()
}

// press dispatches a key press to the current switcher.
func (d *Daemon) press(reverse bool) {
	d.mu.RLock()
	switcher := d.switcher
	d.mu.RUnlock()

	switcher.Press(reverse)
}

// Run starts the keyboard listener and blocks until ctx is cancelled, then shuts
// down cleanly.
func (d *Daemon) Run(ctx context.Context) error {
	defer func() { _ = d.logger.Close() }()

	go d.keyboard.Listen(d.press)

	d.logger.Info("event handler has been initialized")

	<-ctx.Done()

	d.keyboard.Stop()
	d.logger.Info("shutting down")

	return nil
}
