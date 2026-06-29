// Package mainthread runs functions on the process's main thread, which owns the
// AppKit run loop. Some macOS APIs (notably Text Input Sources) abort when called
// from a thread without a running CFRunLoop, so off-main callers must funnel
// through here.
package mainthread

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation

// Implemented in mainthread.m.
void bgkDispatchMain(unsigned long long handle);
*/
import "C"

import "sync"

var (
	mu      sync.Mutex
	lastID  uint64
	pending = make(map[uint64]func())
)

// Run executes fn on the main thread, blocking until fn returns. If called from
// the main thread it runs fn directly.
func Run(fn func()) {
	mu.Lock()
	lastID++
	id := lastID
	pending[id] = fn
	mu.Unlock()

	C.bgkDispatchMain(C.ulonglong(id))

	mu.Lock()
	delete(pending, id)
	mu.Unlock()
}
