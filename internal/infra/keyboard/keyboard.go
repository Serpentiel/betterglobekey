// Package keyboard adapts the global keyboard hook (gohook) into a narrow
// listener for Globe (fn) key releases.
package keyboard

import hook "github.com/robotn/gohook"

// keyFn is the raw key code of the macOS Globe (fn) key.
const keyFn uint16 = 179

// Listener delivers Globe key releases from the global keyboard hook.
type Listener struct{}

// New returns a Listener.
func New() *Listener {
	return &Listener{}
}

// Listen starts the global keyboard hook and invokes onKeyUp for every Globe key
// release. It blocks until Stop is called, at which point the hook is torn down.
func (*Listener) Listen(onKeyUp func()) {
	for event := range hook.Start() {
		if event.Kind == hook.KeyUp && event.Rawcode == keyFn {
			onKeyUp()
		}
	}
}

// Stop tears down the global keyboard hook, unblocking Listen.
func (*Listener) Stop() {
	hook.End()
}
